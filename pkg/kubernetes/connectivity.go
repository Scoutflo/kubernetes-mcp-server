package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// CheckServiceConnectivity attempts to connect to a Kubernetes service
// to verify network connectivity by creating a temporary pod inside the cluster
// and executing a curl command to the service
func (k *Kubernetes) CheckServiceConnectivity(ctx context.Context, serviceName string) (string, error) {
	// Extract hostname and port from service name
	parts := strings.Split(serviceName, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid service name format, expected 'service:port', got: %s", serviceName)
	}

	hostname := parts[0]
	port := parts[1]
	address := fmt.Sprintf("%s:%s", hostname, port)

	// First try regular DNS resolution to see if there might be an issue
	_, err := net.LookupHost(hostname)
	dnsMsg := ""
	if err != nil {
		dnsMsg = fmt.Sprintf("DNS resolution failed: %v\nAttempting in-cluster check...", err)
	}

	// Create a temporary pod in the default namespace to run curl
	podName := fmt.Sprintf("connectivity-test-%d", time.Now().Unix())
	namespace := namespaceOrDefault("")

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "curl",
					Image: "curlimages/curl:latest",
					Command: []string{
						"sleep",
						"120", // Pod will live for 2 minutes max
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	// Create the pod
	_, err = k.clientSet.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create connectivity test pod: %v", err)
	}

	// Set up deferred deletion of the pod
	defer func() {
		deleteCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = k.clientSet.CoreV1().Pods(namespace).Delete(deleteCtx, podName, metav1.DeleteOptions{})
	}()

	// Wait for the pod to be ready
	err = waitForPodRunning(ctx, k.clientSet, namespace, podName, 60*time.Second)
	if err != nil {
		return "", fmt.Errorf("failed waiting for connectivity test pod to start: %v", err)
	}

	// Define the command to execute
	command := []string{"curl", "-v", "-m", "10", address}

	// Execute the command in the pod
	execReq := k.clientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: "curl",
			Command:   command,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	var stdout, stderr bytes.Buffer
	exec, err := remotecommand.NewSPDYExecutor(k.cfg, "POST", execReq.URL())
	if err != nil {
		return "", fmt.Errorf("failed to create exec command: %v", err)
	}

	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	// Build the result message
	result := fmt.Sprintf("In-cluster connectivity check to %s:\n", address)
	if dnsMsg != "" {
		result += dnsMsg + "\n"
	}

	if err != nil {
		result += fmt.Sprintf("Connection check failed: %v\n", err)
		if stderr.Len() > 0 {
			result += fmt.Sprintf("Error output: %s\n", stderr.String())
		}
		return result, nil
	}

	result += "Connection successful from inside the cluster\n"
	if stderr.Len() > 0 {
		result += fmt.Sprintf("Connection details: %s\n", stderr.String())
	}
	if stdout.Len() > 0 {
		result += fmt.Sprintf("Response: %s\n", stdout.String())
	}

	return result, nil
}

// waitForPodRunning waits for a pod to reach Running state
func waitForPodRunning(ctx context.Context, clientset kubernetes.Interface, namespace, podName string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for pod %s to be ready", podName)
		default:
			pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			if pod.Status.Phase == corev1.PodRunning {
				return nil
			}

			// If pod failed or completed, return error
			if pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodSucceeded {
				return fmt.Errorf("pod %s is in phase %s, not Running", podName, pod.Status.Phase)
			}

			time.Sleep(1 * time.Second)
		}
	}
}
