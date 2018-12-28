package worker

import (
	"bufio"
	"log"
	"time"

	"github.com/clockworksoul/cog2/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

type Worker struct {
	CommandParameters []string
	DockerClient      *client.Client
	DockerContext     context.Context
	DockerHost        string
	ExecutionTimeout  time.Duration
	ImageName         string
	done              chan struct{}
}

func NewWorker(imageName string, commandParams ...string) (*Worker, error) {
	dcli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	// Reset the default host
	err = client.WithHost(config.GetDockerConfigs().DockerHost)(dcli)
	if err != nil {
		return nil, err
	}

	return &Worker{
		CommandParameters: commandParams,
		DockerClient:      dcli,
		DockerContext:     context.Background(),
		DockerHost:        config.GetDockerConfigs().DockerHost,
		ExecutionTimeout:  1 * time.Minute,
		ImageName:         imageName,
		done:              make(chan struct{}, 1),
	}, nil
}

// Start triggers a worker to run a container according to its settings.
// It returns a string channel that emits the container's combined stdout and stderr streams.
func (w *Worker) Start() (<-chan string, error) {
	cli := w.DockerClient
	ctx := w.DockerContext
	imageName := w.ImageName
	timeout := w.ExecutionTimeout

	// Start the image pull. This blocks until the pull is complete.
	err := w.pullImage()
	if err != nil {
		return nil, err
	}

	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image: imageName,
			Cmd:   w.CommandParameters,
			Tty:   true,
		},
		nil, nil, "")
	if err != nil {
		return nil, err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	// Begin the timeout counter for this container
	go func() {
		<-time.After(timeout)
		err := w.DockerClient.ContainerStop(w.DockerContext, resp.ID, nil)
		if err != nil {
			log.Printf("Failed to stop container %s: %s", resp.ID, err.Error())
		}
	}()

	// Watch for the container to enter "not running" state. This supports the Stopped() method.
	go func() {
		chwait, _ := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		<-chwait
		w.done <- struct{}{}
	}()

	// Build the channel that will stream back the container logs
	logs, err := w.buildContainerLogChannel(resp.ID)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

// Stopped returns a channel that blocks until this worker's container has stopped.
func (w *Worker) Stopped() <-chan struct{} {
	return w.done
}

// buildContainerLogChannel constructs the log output channel returned by Start()
func (w *Worker) buildContainerLogChannel(containerID string) (<-chan string, error) {
	options := types.ContainerLogsOptions{Follow: true, ShowStdout: true, ShowStderr: true}
	out, err := w.DockerClient.ContainerLogs(w.DockerContext, containerID, options)
	if err != nil {
		return nil, err
	}

	logs := make(chan string)
	go func() {
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			logs <- scanner.Text()
		}

		close(logs)
	}()

	return logs, nil
}

// pullImage pull the worker's image. It blocks until the pull is complete.
func (w *Worker) pullImage() error {
	cli := w.DockerClient
	ctx := w.DockerContext
	imageName := w.ImageName

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

	return nil
}
