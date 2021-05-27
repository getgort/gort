package worker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/clockworksoul/gort/config"
	gcontainer "github.com/clockworksoul/gort/container"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
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
func (w *Worker) Start() (<-chan string, error) {
	// Track time spent in this method
	startTime := time.Now()

	cli := w.DockerClient
	imageName := w.ImageName
	entryPoint := w.EntryPoint

	event := log.
		WithField("image", w.ImageName).
		WithField("entry", entryPoint).
		WithField("params", strings.Join(w.CommandParameters, " "))

	ctx, cancel := context.WithTimeout(context.TODO(), w.ExecutionTimeout)
	defer cancel()

	// Start the image pull. This blocks until the pull is complete.
	err := w.pullImage(false)
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

	resp, err := cli.ContainerCreate(ctx, &cfg, nil, nil, nil, "")
	if err != nil {
		return nil, err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}
	defer func() {
		c, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		duration := 10 * time.Second
		w.DockerClient.ContainerStop(c, resp.ID, &duration)
		w.DockerClient.ContainerRemove(c, resp.ID, types.ContainerRemoveOptions{})
		event.Trace("container stopped and removed")
	}()

	// Watch for the container to enter "not running" state. This supports the Stopped() method.
	go func() {
		chwait, errs := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		event = event.WithField("duration", time.Since(startTime))

		select {
		case ok := <-chwait:
			if ok.Error != nil && ok.Error.Message != "" {
				event = event.WithError(fmt.Errorf(ok.Error.Message))
			}

			event.WithField("status", ok.StatusCode).
				Info("Worker completed")
			w.ExitStatus <- ok.StatusCode

		case err := <-errs:
			event.WithField("duration", time.Since(startTime)).
				WithError(err).
				Error("Error running container")
			w.ExitStatus <- 500
		}
	}()

	// Build the channel that will stream back the container logs
	logs, err := gcontainer.BuildContainerLogChannel(ctx, cli, resp.ID)
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
func (w *Worker) imageExistsLocally(image string) (bool, error) {
	if strings.IndexByte(image, ':') == -1 {
		image += ":latest"
	}

	ctx := context.TODO()
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
func (w *Worker) pullImage(force bool) error {
	cli := w.DockerClient
	ctx := context.TODO()
	imageName := w.ImageName

	exists, err := w.imageExistsLocally(imageName)
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
			Debugf("Container image pulled", imageName, time.Since(startTime))
	}

	return nil
}
