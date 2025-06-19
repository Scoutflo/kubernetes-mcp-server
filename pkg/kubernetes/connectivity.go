package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CheckServiceConnectivity attempts to connect to a Kubernetes service
// to verify network connectivity by creating a temporary pod inside the cluster
// and executing a curl command to the service
func (k *Kubernetes) CheckServiceConnectivity(ctx context.Context, serviceName string) (string, error) {
	// Create query parameters for the API call
	endpoint := fmt.Sprintf("/apis/v1/check-service-connectivity?service_name=%s", url.QueryEscape(serviceName))

	// Make API request to check service connectivity
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to check service connectivity: %w", err)
	}

	// Parse the response to extract the result
	var result struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse connectivity check response: %v", err)
	}

	return result.Result, nil
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

// CheckIngressConnectivity attempts to connect to a Kubernetes ingress host
// to verify network connectivity by creating a temporary pod inside the cluster
// and executing a curl command to the ingress host
func (k *Kubernetes) CheckIngressConnectivity(ctx context.Context, ingressHost string) (string, error) {
	// Create query parameters for the API call
	endpoint := fmt.Sprintf("/apis/v1/check-ingress-connectivity?ingress_host=%s", url.QueryEscape(ingressHost))

	// Make API request to check ingress connectivity
	response, err := k.MakeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to check ingress connectivity: %w", err)
	}

	// Parse the response to extract the result
	var result struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse ingress connectivity check response: %v", err)
	}

	return result.Result, nil
}
