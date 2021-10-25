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

package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8srest "k8s.io/client-go/rest"
)

// Worker represents a container executor. It has a lifetime of a single command execution.
type KubernetesWorker struct {
	clientset         *kubernetes.Clientset
	command           data.CommandRequest
	commandParameters []string
	entryPoint        []string
	exitStatus        chan int64
	imageName         string
	token             rest.Token
}

// New will build and returns a new Worker for a single command execution.
func New(command data.CommandRequest, token rest.Token) (*KubernetesWorker, error) {
	image := command.Bundle.Docker.Image
	tag := command.Bundle.Docker.Tag
	entrypoint := command.Command.Executable
	params := command.Parameters

	if tag == "" {
		tag = "latest"
	}

	// creates the in-cluster config
	config, err := k8srest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &KubernetesWorker{
		clientset:         clientset,
		command:           command,
		commandParameters: params,
		entryPoint:        entrypoint,
		exitStatus:        make(chan int64, 1),
		imageName:         image + ":" + tag,
		token:             token,
	}, nil
}

// Start triggers a worker to run a container according to its settings.
// It returns a string channel that emits the container's combined stdout and stderr streams.
func (w *KubernetesWorker) Start(ctx context.Context) (<-chan string, error) {
	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	pods, err := w.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	logs := make(chan string, 1)

	j := &batchv1.Job{}

	w.clientset.BatchV1().Jobs("").Create(ctx, j, metav1.CreateOptions{})

	go func() {
		logs <- fmt.Sprintf("There are %d pods in the cluster.", len(pods.Items))
		close(logs)

		w.exitStatus <- 0
		close(w.exitStatus)
	}()

	return logs, nil
}

// Cleanup will stop (if it's not already stopped) a worker process and cleanup
// any resources it's using. If the worker fails to stop gracefully within a
// timeframe specified by the timeout argument, it is forcefully terminated
// (killed). If the timeout is nil, the engine's default is used. A negative
// timeout indicates no timeout: no forceful termination is performed.
func (w *KubernetesWorker) Stop(ctx context.Context, timeout *time.Duration) {
}

// Stopped returns a channel that blocks until this worker's container has stopped.
// The value emitted is the exit status code of the underlying process.
func (w *KubernetesWorker) Stopped() <-chan int64 {
	return w.exitStatus
}
