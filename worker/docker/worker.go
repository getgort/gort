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

package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/getgort/gort/config"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/telemetry"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// Worker represents a container executor. It has a lifetime of a single command execution.
type ContainerWorker struct {
	command           data.CommandRequest
	commandParameters []string
	configs           map[string]string
	containerID       string
	dockerClient      *client.Client
	dockerHost        string
	entryPoint        []string
	exitStatus        chan int64
	imageName         string
	token             rest.Token
}

// New will build and returns a new Worker for a single command execution.
func New(command data.CommandRequest, token rest.Token) (*ContainerWorker, error) {
	entrypoint := command.Command.Executable
	params := command.Parameters

	dcli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	// Reset the default host
	err = client.WithHost(config.GetDockerConfigs().DockerHost)(dcli)
	if err != nil {
		return nil, err
	}

	return &ContainerWorker{
		command:           command,
		commandParameters: params,
		configs:           map[string]string{},
		dockerClient:      dcli,
		dockerHost:        config.GetDockerConfigs().DockerHost,
		entryPoint:        entrypoint,
		exitStatus:        make(chan int64),
		imageName:         command.Bundle.ImageFull(),
		token:             token,
	}, nil
}

func (w *ContainerWorker) Initialize(dc []data.DynamicConfiguration) {
	for _, c := range dc {
		w.configs[c.Key] = c.Value
	}
}

// Start triggers a worker to run a container according to its settings.
// It returns a string channel that emits the container's combined stdout and stderr streams.
func (w *ContainerWorker) Start(ctx context.Context) (<-chan string, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "worker.docker.Start")
	defer sp.End()

	// Track time spent in this method
	startTime := time.Now()

	cli := w.dockerClient
	imageName := w.imageName
	entryPoint := w.entryPoint

	event := log.
		WithField("image", w.imageName).
		WithField("entry", entryPoint).
		WithField("params", strings.Join(w.commandParameters, " "))
	if sp.SpanContext().HasTraceID() {
		event = event.WithField("trace.id", sp.SpanContext().TraceID())
	}

	sp.SetAttributes(
		attribute.String("image", w.imageName),
		attribute.String("entry", strings.Join(entryPoint, " ")),
		attribute.String("params", strings.Join(w.commandParameters, " ")),
	)

	// Start the image pull. This blocks until the pull is complete.
	err := w.pullImage(ctx, false)
	if err != nil {
		return nil, err
	}

	cfg := container.Config{
		Image: imageName,
		Cmd:   w.commandParameters,
		Tty:   true,
		Env:   w.envVars(),
	}

	if len(entryPoint) > 0 {
		cfg.Entrypoint = entryPoint
	}

	// Create the container
	resp, err := func() (container.ContainerCreateCreatedBody, error) {
		ctx, sp := tr.Start(ctx, "worker.docker.ContainerCreate")
		defer sp.End()

		// If a host network is defined, set it here.
		var hc *container.HostConfig
		if network := config.GetDockerConfigs().Network; network != "" {
			hc = &container.HostConfig{NetworkMode: container.NetworkMode(network)}
		}

		return cli.ContainerCreate(ctx, &cfg, hc, nil, nil, "")
	}()
	if err != nil {
		return nil, err
	}

	w.containerID = resp.ID
	event = event.WithField("containerID", w.containerID)

	// Start the container
	err = func() error {
		ctx, sp := tr.Start(ctx, "worker.docker.ContainerStart")
		defer sp.End()
		return cli.ContainerStart(ctx, w.containerID, types.ContainerStartOptions{})
	}()
	if err != nil {
		return nil, err
	}

	// Watch for the container to enter "not running" state. This supports the Stopped() method.
	go func() {
		chwait, errs := cli.ContainerWait(ctx, w.containerID, container.WaitConditionNotRunning)
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

		w.exitStatus <- status
		sp.SetAttributes(attribute.Int64("status", status))
	}()

	// Build the channel that will stream back the container logs.
	// Blocks until the container stops.
	logs, err := BuildContainerLogChannel(ctx, cli, w.containerID)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

// Stop will stop (if it's not already stopped) a worker process and clean up
// any resources it's using. If the worker fails to stop gracefully within a
// timeframe specified by the timeout argument, it is forcefully terminated
// (killed). If the timeout is nil, the engine's default is used. A negative
// timeout indicates no timeout: no forceful termination is performed.
func (w *ContainerWorker) Stop(ctx context.Context, timeout *time.Duration) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "worker.docker.Stop")
	defer sp.End()

	func() error {
		ctx, sp := tr.Start(ctx, "docker.ContainerStop")
		defer sp.End()
		return w.dockerClient.ContainerStop(ctx, w.containerID, timeout)
	}()

	func() error {
		ctx, sp := tr.Start(ctx, "docker.ContainerRemove")
		defer sp.End()
		return w.dockerClient.ContainerRemove(ctx, w.containerID, types.ContainerRemoveOptions{})
	}()

	log.WithField("containerID", w.containerID).Trace("container stopped and removed")
}

// Stopped returns a channel that blocks until this worker's container has stopped.
// The value emitted is the exit status code of the underlying process.
func (w *ContainerWorker) Stopped() <-chan int64 {
	return w.exitStatus
}

func (w *ContainerWorker) envVars() []string {
	var env []string

	for k, v := range w.configs {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	vars := map[string]string{
		`GORT_ADAPTER`:       w.command.Adapter,
		`GORT_BUNDLE`:        w.command.Bundle.Name,
		`GORT_COMMAND`:       w.command.Command.Name,
		`GORT_CHAT_ID`:       w.command.UserID,
		`GORT_INVOCATION_ID`: fmt.Sprintf("%d", w.command.RequestID),
		`GORT_ROOM`:          w.command.ChannelID,
		`GORT_SERVICE_TOKEN`: w.token.Token,
		`GORT_SERVICES_ROOT`: config.GetGortServerConfigs().APIURLBase,
		`GORT_USER`:          w.command.UserName,
	}

	for k, v := range vars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env
}

// imageExistsLocally returns true if the specified image is present and
// accessible to the docker daemon.
func (w *ContainerWorker) imageExistsLocally(ctx context.Context, image string) (bool, error) {
	if strings.IndexByte(image, ':') == -1 {
		image += ":latest"
	}

	images, err := w.dockerClient.ImageList(ctx, types.ImageListOptions{})
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
func (w *ContainerWorker) pullImage(ctx context.Context, force bool) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "worker.docker.pullImage")
	defer sp.End()

	cli := w.dockerClient
	imageName := w.imageName

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
