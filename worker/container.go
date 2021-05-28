package worker

import (
	"bufio"
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// BuildContainerLogChannel accepts a pointer to a Docker client.Client and a
// container ID and returns a string channel that provides all events from the
// container's standard output and error streams.
func BuildContainerLogChannel(ctx context.Context, client *client.Client, containerID string) (<-chan string, error) {
	options := types.ContainerLogsOptions{Follow: true, ShowStdout: true, ShowStderr: true}
	out, err := client.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return nil, err
	}

	logs := wrapReaderInChannel(out)

	return logs, nil
}

// BuildContainerLogChannels accepts a pointer to a Docker client.Client and a
// container ID and returns tww string channels, one that provides all events
// emitted by the container's standard output stream, and a second for standard
// error.
func BuildContainerLogChannels(ctx context.Context, client *client.Client, containerID string) (stdout, stderr <-chan string, err error) {
	var outr, errr io.ReadCloser

	outr, err = client.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{Follow: true, ShowStdout: true})
	if err != nil {
		return
	}

	errr, err = client.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{Follow: true, ShowStdout: true})
	if err != nil {
		return
	}

	stdout = wrapReaderInChannel(outr)
	stderr = wrapReaderInChannel(errr)

	return
}

func wrapReaderInChannel(rc io.Reader) <-chan string {
	ch := make(chan string)

	go func() {
		scanner := bufio.NewScanner(rc)
		for scanner.Scan() {
			ch <- scanner.Text()
		}

		close(ch)
	}()

	return ch
}
