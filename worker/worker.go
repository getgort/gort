/*
 * Copyright 2021 The Gort Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package worker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/getgort/gort/config"
	"github.com/getgort/gort/telemetry"
)

// Worker represents a container executor. It has a lifetime of a single command execution.
type Worker struct {
	CommandParameters []string
	DockerClient      *client.Client
	DockerHost        string
	EntryPoint        string
	ExitStatus        chan int64
	ExecutionTimeout  time.Duration
	ImageName         string
}

// NewWorker will build and returns a new Worker for a single command execution.
func NewWorker(image string, tag string, entryPoint string, commandParams ...string) (*Worker, error) {
	dcli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	// Reset the default host
	err = client.WithHost(config.GetDockerConfigs().DockerHost)(dcli)
	if err != nil {
		return nil, err
	}

	if tag == "" {
		tag = "latest"
	}

	return &Worker{
		CommandParameters: commandParams,
		DockerClient:      dcli,
		DockerHost:        config.GetDockerConfigs().DockerHost,
		EntryPoint:        entryPoint,
		ExecutionTimeout:  1 * time.Minute,
		ImageName:         image + ":" + tag,
		ExitStatus:        make(chan int64),
	}, nil
}

// Start triggers a worker to run a container according to its settings.
// It returns a string channel that emits the container's combined stdout and stderr streams.
func (w *Worker) Start(ctx context.Context) (<-chan string, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "worker.Start")
	defer sp.End()

	// Track time spent in this method
	startTime := time.Now()

	cli := w.DockerClient
	imageName := w.ImageName
	entryPoint := w.EntryPoint

	event := log.
		WithField("image", w.ImageName).
		WithField("entry", entryPoint).
		WithField("params", strings.Join(w.CommandParameters, " "))
	if sp.SpanContext().HasTraceID() {
		event = event.WithField("trace.id", sp.SpanContext().TraceID())
	}

	sp.SetAttributes(
		attribute.String("image", w.ImageName),
		attribute.String("entry", entryPoint),
		attribute.String("params", strings.Join(w.CommandParameters, " ")),
	)

	ctx, cancel := context.WithTimeout(ctx, w.ExecutionTimeout)
	defer cancel()

	// Start the image pull. This blocks until the pull is complete.
	err := w.pullImage(ctx, false)
	if err != nil {
		return nil, err
	}

	cfg := container.Config{
		Image: imageName,
		Cmd:   w.CommandParameters,
		Tty:   true,
	}

	if entryPoint != "" {
		cfg.Entrypoint = []string{entryPoint}
	}

	// Create the container
	resp, err := func() (container.ContainerCreateCreatedBody, error) {
		ctx, sp := tr.Start(ctx, "docker.ContainerCreate")
		defer sp.End()
		return cli.ContainerCreate(ctx, &cfg, nil, nil, nil, "")
	}()
	if err != nil {
		return nil, err
	}

	// Start the container
	err = func() error {
		ctx, sp := tr.Start(ctx, "docker.ContainerStart")
		defer sp.End()
		return cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	}()
	if err != nil {
		return nil, err
	}

	// Make sure the container is cleaned up when we're done
	defer func() {
		ctx, sp := tr.Start(ctx, "docker.ContainerCleanup")
		defer sp.End()

		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		duration := 10 * time.Second
		w.DockerClient.ContainerStop(ctx, resp.ID, &duration)
		w.DockerClient.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
		event.Trace("container stopped and removed")
	}()

	// Watch for the container to enter "not running" state. This supports the Stopped() method.
	go func() {
		chwait, errs := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		event = event.WithField("duration", time.Since(startTime))

		var status int64

		select {
		case ok := <-chwait:
			if ok.Error != nil && ok.Error.Message != "" {
				event = event.WithError(fmt.Errorf(ok.Error.Message))
				sp.SetAttributes(attribute.String("error", ok.Error.Message))
				telemetry.Errors().WithError(fmt.Errorf(ok.Error.Message)).Commit(ctx)
			}

			status = ok.StatusCode
			event.WithField("status", status).
				Info("Worker completed")

		case err := <-errs:
			status = 500
			event.WithField("status", status).
				WithError(err).
				Error("Error running container")
			sp.SetAttributes(attribute.String("error", err.Error()))
		}

		w.ExitStatus <- status
		sp.SetAttributes(attribute.Int64("status", status))
	}()

	// Build the channel that will stream back the container logs.
	// Blocks until the container stops.
	logs, err := BuildContainerLogChannel(ctx, cli, resp.ID)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

// Stopped returns a channel that blocks until this worker's container has stopped.
// The value emitted is the exit status code of the underlying process.
func (w *Worker) Stopped() <-chan int64 {
	return w.ExitStatus
}

// imageExistsLocally returns true if the specified image is present and
// accessible to the docker daemon.
func (w *Worker) imageExistsLocally(ctx context.Context, image string) (bool, error) {
	if strings.IndexByte(image, ':') == -1 {
		image += ":latest"
	}

	images, err := w.DockerClient.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return false, err
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if image == tag {
				return true, nil
			}
		}
	}

	return false, nil
}

// pullImage pull the worker's image. It blocks until the pull is complete.
func (w *Worker) pullImage(ctx context.Context, force bool) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "worker.pullImage")
	defer sp.End()

	cli := w.DockerClient
	imageName := w.ImageName

	exists, err := w.imageExistsLocally(ctx, imageName)
	if err != nil {
		return err
	}

	if force || !exists {
		startTime := time.Now()

		log.WithField("image", imageName).Trace("Pulling container image", imageName)

		reader, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
		if err != nil {
			return err
		}
		defer reader.Close()

		// Watch the daemon output until we get an EOF
		bytes := make([]byte, 256)
		var e error
		for e == nil {
			_, e = reader.Read(bytes)
		}

		log.WithField("image", imageName).
			WithField("duration", time.Since(startTime)).
			Debug("Container image pulled")
	}

	return nil
}
