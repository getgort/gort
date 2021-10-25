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
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"

	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8srest "k8s.io/client-go/rest"
)

const namespace = "default"

// KubernetesWorker represents a container executor. It has a lifetime of a single command execution.
type KubernetesWorker struct {
	clientset         *kubernetes.Clientset
	command           data.CommandRequest
	commandParameters []string
	entryPoint        []string
	exitStatus        chan int64
	imageName         string
	jobName           string
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
	var backoffLimit int32 = 0

	envVars, err := w.envVars(ctx)
	if err != nil {
		return nil, err
	}

	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s.%s", w.command.Bundle.Name, w.command.Command.Name),
			Labels: map[string]string{
				"gort.bundle":  w.command.Bundle.Name,
				"gort.command": w.command.Command.Name,
				"gort.request": fmt.Sprintf("%d", w.command.RequestID),
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "command",
							Image:   w.imageName,
							Command: w.entryPoint,
							Args:    w.commandParameters,
							Env:     envVars,
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	if len(w.entryPoint) > 0 {
		job.Spec.Template.Spec.Containers[0].Command = w.entryPoint
	}

	jobInterface := w.clientset.BatchV1().Jobs(namespace)
	job, err = jobInterface.Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start kubernetes job: %w", err)
	}

	w.jobName = job.Name

	watchInterface, err := jobInterface.Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	go func() {
		events := watchInterface.ResultChan()
		complete := false

		for !complete {
			select {
			case e := <-events:
				if j, ok := e.Object.(*batchv1.Job); ok && j.Name == w.jobName {
					for _, c := range j.Status.Conditions {
						switch c.Type {
						case batchv1.JobComplete:
							fallthrough
						case batchv1.JobFailed:
							complete = true
						}
					}
				}
			case <-ctx.Done():
				complete = true
			}
		}

		w.exitStatus <- int64(job.Status.Failed)
	}()

	// Build the channel that will stream back the job logs.
	// Blocks until the container stops.
	logs, err := w.getJobLogs(ctx)
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
func (w *KubernetesWorker) Stop(ctx context.Context, timeout *time.Duration) {
	jobInterface := w.clientset.BatchV1().Jobs(namespace)

	deleteOptions := metav1.DeleteOptions{}
	if timeout != nil && timeout.Seconds() > 0 {
		seconds := int64(timeout.Seconds())
		deleteOptions.GracePeriodSeconds = &seconds
	}

	if err := jobInterface.Delete(ctx, w.jobName, deleteOptions); err != nil {
		log.WithError(err).WithField("jobName", w.jobName).Trace("failed to delete job")
		return
	}

	log.WithField("jobName", w.jobName).Trace("job stopped and removed")
}

// Stopped returns a channel that blocks until this worker's container has stopped.
// The value emitted is the exit status code of the underlying process.
func (w *KubernetesWorker) Stopped() <-chan int64 {
	return w.exitStatus
}

func (w *KubernetesWorker) envVars(ctx context.Context) ([]corev1.EnvVar, error) {
	gortIP, gortPort, err := w.findGortEndpoint(ctx)
	if err != nil {
		return nil, err
	}

	vars := map[string]string{
		`GORT_BUNDLE`:        w.command.Bundle.Name,
		`GORT_COMMAND`:       w.command.Command.Name,
		`GORT_CHAT_HANDLE`:   w.command.UserID,
		`GORT_INVOCATION_ID`: fmt.Sprintf("%d", w.command.RequestID),
		`GORT_ROOM`:          w.command.ChannelID,
		`GORT_SERVICE_TOKEN`: w.token.Token,
		`GORT_SERVICES_ROOT`: fmt.Sprintf("http://%s:%d", gortIP, gortPort),
	}

	var env []corev1.EnvVar

	for k, v := range vars {
		env = append(env, corev1.EnvVar{Name: k, Value: v})
	}

	return env, nil
}

func (w *KubernetesWorker) findGortEndpoint(ctx context.Context) (string, int32, error) {
	epInterface := w.clientset.CoreV1().Endpoints(namespace)

	ep, err := epInterface.Get(ctx, "gort", metav1.GetOptions{})
	if err != nil {
		return "", 0, fmt.Errorf("failed to get gort endpoint: %w", err)
	}

	return ep.Subsets[0].Addresses[0].IP, ep.Subsets[0].Ports[0].Port, nil
}

func (w *KubernetesWorker) getJobLogs(ctx context.Context) (<-chan string, error) {
	var pod *corev1.Pod

	podInterface := w.clientset.CoreV1().Pods(namespace)

	// listOptions := metav1.ListOptions{LabelSelector: "job-name=" + w.jobName}
	listOptions := metav1.ListOptions{}

	wi, err := podInterface.Watch(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed watch pod list: %w", err)
	}

	resultChan := wi.ResultChan()

	for pod == nil {
		select {
		case e := <-resultChan:
			if p, ok := e.Object.(*corev1.Pod); ok && strings.HasPrefix(p.Name, w.jobName) {
				if p.Status.Phase != "Pending" {
					pod = p
				}
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	podLogOpts := &corev1.PodLogOptions{}
	req := podInterface.GetLogs(pod.Name, podLogOpts)

	podLogs, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to stream logs from pod: %w", err)
	}

	return wrapReaderInChannel(podLogs), nil
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
			log.WithError(err).Error("error scanning reader")
		}

		close(ch)
	}()

	return ch
}
