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
	"time"

	"github.com/getgort/gort/config"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/worker/docker"
	"github.com/getgort/gort/worker/kubernetes"
)

// Worker represents a container executor. It has a lifetime of a single command execution.
type Worker interface {
	Initialize([]data.DynamicConfiguration)
	Start(ctx context.Context) (<-chan string, error)
	Stop(ctx context.Context, timeout *time.Duration)
	Stopped() <-chan int64
}

// New will build and return a new Worker for a single command execution.
func New(command data.CommandRequest, token rest.Token) (Worker, error) {
	dockerDefined := !config.Undefined(config.GetDockerConfigs())
	kubernetesDefined := !config.Undefined(config.GetKubernetesConfigs())

	switch {
	case dockerDefined && kubernetesDefined:
		return nil, fmt.Errorf("exactly one of the following config sections expected: docker, kubernetes")
	case dockerDefined:
		return docker.New(command, token)
	case kubernetesDefined:
		return kubernetes.New(command, token)
	default:
		return nil, fmt.Errorf("exactly one of the following config sections expected: docker, kubernetes")
	}
}
