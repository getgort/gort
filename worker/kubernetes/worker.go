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
	"os"
	"time"

	"github.com/getgort/gort/config"
	"github.com/getgort/gort/data"
	"github.com/getgort/gort/data/rest"
	"github.com/getgort/gort/telemetry"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8srest "k8s.io/client-go/rest"
)

// KubernetesWorker represents a container executor. It has a lifetime of a single command execution.
type KubernetesWorker struct {
	clientset         *kubernetes.Clientset
	command           data.CommandRequest
	commandParameters []string
	entryPoint        []string
	exitStatus        chan int64
	imageName         string
	jobName           string
	namespace         string
	token             rest.Token
}

// New will build and returns a new Worker for a single command execution.
func New(command data.CommandRequest, token rest.Token) (*KubernetesWorker, error) {
	entrypoint := command.Command.Executable
	params := command.Parameters

	// creates the in-cluster config
	kconfig, err := k8srest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(kconfig)
	if err != nil {
		return nil, err
	}

	w := &KubernetesWorker{
		clientset:         clientset,
		command:           command,
		commandParameters: params,
		entryPoint:        entrypoint,
		exitStatus:        make(chan int64, 1),
		imageName:         command.Bundle.ImageFull(),
		token:             token,
	}

	return w, nil
}

// Start triggers a worker to run a Kubernetes Job. It returns a string
// channel that emits the container's combined stdout and stderr streams.
func (w *KubernetesWorker) Start(ctx context.Context) (<-chan string, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "worker.kubernetes.Start")
	defer sp.End()

	// We have to set the namespace!
	pod, err := w.findGortPod(ctx)
	if err != nil {
		return nil, err
	}
	w.namespace = pod.Namespace

	job, err := w.buildJobData(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build job struct: %w", err)
	}

	jobInterface := w.clientset.BatchV1().Jobs(w.namespace)
	job, err = jobInterface.Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start kubernetes job: %w", err)
	}

	w.jobName = job.Name

	// Watch the Job's pod for termination.
	if err := w.watchForPodTermination(ctx); err != nil {
		return nil, fmt.Errorf("failed to watch job pod: %w", err)
	}

	// Build the channel that will stream back the job logs.
	// Blocks until the container stops.
	logs, err := w.getJobOutput(ctx)
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
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "worker.kubernetes.Stop")
	defer sp.End()

	// Clean up the Job resource
	err := func() error {
		ctx, sp := tr.Start(ctx, "worker.kubernetes.Stop.CleanUpJob")
		defer sp.End()

		// Clean up the job
		jobInterface := w.clientset.BatchV1().Jobs(w.namespace)
		deleteOptions := metav1.DeleteOptions{}
		if timeout != nil && timeout.Seconds() > 0 {
			seconds := int64(timeout.Seconds())
			deleteOptions.GracePeriodSeconds = &seconds
		}

		return jobInterface.Delete(ctx, w.jobName, deleteOptions)
	}()
	if err != nil {
		log.WithError(err).WithField("jobName", w.jobName).Error("Failed to delete job")
		return
	}

	// Clean up the Job's Pod resource
	err = func() error {
		ctx, sp := tr.Start(ctx, "worker.kubernetes.Stop.CleanUpPod")
		defer sp.End()

		podInterface := w.clientset.CoreV1().Pods(w.namespace)
		listOptions := metav1.ListOptions{LabelSelector: "job-name=" + w.jobName}
		pl, err := podInterface.List(ctx, listOptions)
		if err != nil {
			return err
		}

		if len(pl.Items) == 0 {
			return fmt.Errorf("failed to find pod for job-name=%s", w.jobName)
		}

		for _, p := range pl.Items {
			if err := podInterface.Delete(ctx, p.Name, metav1.DeleteOptions{}); err != nil {
				return err
			}
		}
		return nil
	}()
	if err != nil {
		log.WithError(err).WithField("jobName", w.jobName).Error("Failed to delete pod")
		return
	}

	log.WithField("jobName", w.jobName).Info("Job stopped and removed")
}

// Stopped returns a channel that blocks until this worker's container has stopped.
// The value emitted is the exit status code of the underlying process.
func (w *KubernetesWorker) Stopped() <-chan int64 {
	return w.exitStatus
}

// Start triggers a worker to run a container according to its settings.
// It returns a string channel that emits the container's combined stdout and stderr streams.
func (w *KubernetesWorker) buildJobData(ctx context.Context) (*batchv1.Job, error) {
	var backoffLimit int32 = 0

	envVars, err := w.envVars(ctx)
	if err != nil {
		return nil, err
	}

	secretEnv := []corev1.EnvFromSource{}

	if w.command.Bundle.Kubernetes.EnvSecret != "" {
		secretEnv = append(secretEnv, corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: w.command.Bundle.Kubernetes.EnvSecret,
				},
			},
		},
		)
	}

	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{APIVersion: "batch/v1", Kind: "Job"},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s.%s-", w.command.Bundle.Name, w.command.Command.Name),
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
					ServiceAccountName: w.command.Bundle.Kubernetes.ServiceAccountName,
					Containers: []corev1.Container{
						{
							Name:    "command",
							Image:   w.imageName,
							Command: w.entryPoint,
							Args:    w.commandParameters,
							Env:     envVars,
							EnvFrom: secretEnv,
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

	return job, nil
}

// envVars builds the default environment variables that get injected into
// the command pod.
func (w *KubernetesWorker) envVars(ctx context.Context) ([]corev1.EnvVar, error) {
	gortIP, gortPort, err := w.findGortEndpoint(ctx)
	if err != nil {
		return nil, err
	}

	vars := map[string]string{
		`GORT_ADAPTER`:       w.command.Adapter,
		`GORT_BUNDLE`:        w.command.Bundle.Name,
		`GORT_COMMAND`:       w.command.Command.Name,
		`GORT_CHAT_ID`:       w.command.UserID,
		`GORT_INVOCATION_ID`: fmt.Sprintf("%d", w.command.RequestID),
		`GORT_ROOM`:          w.command.ChannelID,
		`GORT_SERVICE_TOKEN`: w.token.Token,
		`GORT_SERVICES_ROOT`: fmt.Sprintf("%s:%d", gortIP, gortPort),
		`GORT_USER`:          w.command.UserName,
	}

	var env []corev1.EnvVar

	for k, v := range vars {
		env = append(env, corev1.EnvVar{Name: k, Value: v})
	}

	return env, nil
}

// findGortEndpoint uses the Kubernetes API to look for Gort's API endpoint.
// It will return an error if it doesn't have permission to "get" endpoint
// resources in the active namespace.
func (w *KubernetesWorker) findGortEndpoint(ctx context.Context) (string, int32, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "worker.kubernetes.findGortEndpoint")
	defer sp.End()

	epInterface := w.clientset.CoreV1().Endpoints(w.namespace)
	fieldSelector := config.GetKubernetesConfigs().EndpointFieldSelector
	labelSelector := config.GetKubernetesConfigs().EndpointFieldSelector

	if fieldSelector == "" && labelSelector == "" {
		labelSelector = "app=gort"
	}

	opts := metav1.ListOptions{FieldSelector: fieldSelector, LabelSelector: labelSelector}
	list, err := epInterface.List(ctx, opts)
	if err != nil {
		return "", 0, fmt.Errorf("failed to list gort endpoints: %w", err)
	}

	switch len(list.Items) {
	case 0:
		return "", 0, fmt.Errorf("failed to find Gort endpoint (fieldSelector=%q labelSelector=%q)", labelSelector, fieldSelector)
	case 1:
		subset := list.Items[0].Subsets[0]
		return subset.Addresses[0].IP, subset.Ports[0].Port, nil
	default:
		return "", 0, fmt.Errorf("found too many endpoints (n=%d fieldSelector=%q labelSelector=%q)", len(list.Items), labelSelector, fieldSelector)
	}
}

// findGortPod attempts to find the Gort service pod.
func (w *KubernetesWorker) findGortPod(ctx context.Context) (*corev1.Pod, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "worker.kubernetes.findGortPod")
	defer sp.End()

	name := os.Getenv("GORT_POD_NAME")

	// The first time this is used, w.namespace may not be defined yet.
	namespace := w.namespace
	if namespace == "" {
		namespace = os.Getenv("GORT_POD_NAMESPACE")
	}

	podInterface := w.clientset.CoreV1().Pods(namespace)

	// If we know the namespace and name, this gets easy.
	if name != "" && namespace != "" {
		pod, err := podInterface.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get pod %s.%s: %w", namespace, name, err)
		}
		return pod, err
	}

	// If we don't know the namespace, we have scan the list of Pods.
	var listOptions metav1.ListOptions

	fieldSelector := config.GetKubernetesConfigs().PodFieldSelector
	labelSelector := config.GetKubernetesConfigs().PodFieldSelector

	// To find the Gort pod, we use three possible strategies:
	// 1: If the Downward API provided the Pod name, we use use that
	// 2: If the config includes selectors, we use those
	// 3: We use the label selector "app=gort" and hope for the best
	if name != "" {
		listOptions = metav1.ListOptions{FieldSelector: "metadata.name=" + name}
	} else if fieldSelector != "" || labelSelector != "" {
		listOptions = metav1.ListOptions{FieldSelector: fieldSelector, LabelSelector: labelSelector}
	} else {
		listOptions = metav1.ListOptions{LabelSelector: "app=gort"}
	}

	list, err := podInterface.List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list gort pods: %w", err)
	}

	switch len(list.Items) {
	case 0:
		return nil, fmt.Errorf("failed to find Gort endpoint (fieldSelector=%q labelSelector=%q)", labelSelector, fieldSelector)
	case 1:
		return &list.Items[0], nil
	default:
		return nil, fmt.Errorf("found too many endpoints (n=%d fieldSelector=%q labelSelector=%q)", len(list.Items), labelSelector, fieldSelector)
	}
}

// getJobOutput returns a channel containing the job's pod's stdout output
// (i.e., its logs). This may pause briefly because it has to wait for the pod
// to exit the Pending state to access its logs. This will return an error if
// it doesn't have "get" and "watch" permissions on "pods/log".
func (w *KubernetesWorker) getJobOutput(ctx context.Context) (<-chan string, error) {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	ctx, sp := tr.Start(ctx, "worker.kubernetes.getJobOutput")
	defer sp.End()

	var pod *corev1.Pod

	listOptions := metav1.ListOptions{LabelSelector: "job-name=" + w.jobName}
	podInterface := w.clientset.CoreV1().Pods(w.namespace)
	wi, err := podInterface.Watch(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to watch pod list: %w", err)
	}

	resultChan := wi.ResultChan()

	for pod == nil {
		select {
		case e := <-resultChan:
			if p, ok := e.Object.(*corev1.Pod); ok {
				if p.Status.Phase != "Pending" {
					pod = p
				}
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	podLogOpts := &corev1.PodLogOptions{Container: "command", Follow: true}
	req := podInterface.GetLogs(pod.Name, podLogOpts)

	podLogs, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to stream logs from pod: %w", err)
	}

	return wrapReaderInChannel(podLogs), nil
}

// watchForPodTermination watches for changes in the job's pod. When its
// container process terminates, it sends the exit code to w.exitStatus.
// An error is returned if the current service account lacks permissions to
// watch on pods in the working namespace.
func (w *KubernetesWorker) watchForPodTermination(ctx context.Context) error {
	tr := otel.GetTracerProvider().Tracer(telemetry.ServiceName)
	_, sp := tr.Start(ctx, "worker.kubernetes.watchForPodTermination")
	defer sp.End()

	listOptions := metav1.ListOptions{LabelSelector: "job-name=" + w.jobName}
	podInterface := w.clientset.CoreV1().Pods(w.namespace)
	watchInterface, err := podInterface.Watch(ctx, listOptions)
	if err != nil {
		return err
	}

	go func() {
		events := watchInterface.ResultChan()

		var terminated *corev1.ContainerStateTerminated
		var complete bool

		for !complete {
			select {
			case e := <-events:
				if pod, ok := e.Object.(*corev1.Pod); ok {
					if len(pod.Status.ContainerStatuses) == 0 {
						continue
					}

					if terminated = pod.Status.ContainerStatuses[0].State.Terminated; terminated == nil {
						continue
					}

					complete = true
				}
			case <-ctx.Done():
				complete = true
			}
		}

		w.exitStatus <- int64(terminated.ExitCode)
	}()

	return nil
}

func wrapReaderInChannel(rc io.Reader) <-chan string {
	ch := make(chan string)
	errs := make(chan error)

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
		close(errs)
	}()

	return ch
}
