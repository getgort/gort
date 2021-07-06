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
	"bufio"
	"context"
	"io"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/getgort/gort/telemetry"
)

// BuildContainerLogChannel accepts a pointer to a Docker client.Client and a
// container ID and returns a string channel that provides all events from the
// container's standard output and error streams.
func BuildContainerLogChannel(ctx context.Context, client *client.Client, containerID string) (<-chan string, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "container.BuildContainerLogChannel")
	defer sp.End()

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
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "container.BuildContainerLogChannels")
	defer sp.End()

	var outRc, errRc io.ReadCloser

	outRc, err = client.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{Follow: true, ShowStdout: true})
	if err != nil {
		return
	}

	errRc, err = client.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{Follow: true, ShowStdout: true})
	if err != nil {
		return
	}

	stdout = wrapReaderInChannel(outRc)
	stderr = wrapReaderInChannel(errRc)

	return
}

func wrapReaderInChannel(rc io.Reader) <-chan string {
	ch := make(chan string)

	go func() {
		scanner := bufio.NewScanner(rc)
		for scanner.Scan() {
			ch <- scanner.Text()
		}

		err := scanner.Err()
		if err != nil && err != io.EOF {
			log.WithError(err).Error("Error scanning reader")
		}

		close(ch)
	}()

	return ch
}
