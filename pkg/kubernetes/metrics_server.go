package kubernetes

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

// metricsClient is a client for the metrics server
var metricsClient *metrics.Clientset

// initMetricsClient initializes the metrics client using the current config
func (k *Kubernetes) initMetricsClient() error {
	var err error
	metricsClient, err = metrics.NewForConfig(k.cfg)
	return err
}

// GetNodeMetrics returns CPU and memory metrics for all nodes or a specific node
func (k *Kubernetes) GetNodeMetrics(ctx context.Context, nodeName string) (string, error) {
	if metricsClient == nil {
		if err := k.initMetricsClient(); err != nil {
			return "", fmt.Errorf("failed to initialize metrics client: %v", err)
		}
	}

	var nodeMetrics *metricsv1beta1.NodeMetricsList
	var err error

	if nodeName != "" {
		// Get metrics for a specific node
		metric, err := metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, nodeName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		nodeMetrics = &metricsv1beta1.NodeMetricsList{
			Items: []metricsv1beta1.NodeMetrics{*metric},
		}
	} else {
		// Get metrics for all nodes
		nodeMetrics, err = metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
		if err != nil {
			return "", err
		}
	}

	return marshal(nodeMetrics)
}

// GetPodMetrics returns CPU and memory metrics for pods in a namespace
func (k *Kubernetes) GetPodMetrics(ctx context.Context, namespace string, podName string) (string, error) {
	if metricsClient == nil {
		if err := k.initMetricsClient(); err != nil {
			return "", fmt.Errorf("failed to initialize metrics client: %v", err)
		}
	}

	if podName != "" {
		// Get metrics for a specific pod
		podMetric, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		podMetrics := &metricsv1beta1.PodMetricsList{
			Items: []metricsv1beta1.PodMetrics{*podMetric},
		}
		return marshal(podMetrics)
	}

	// Get metrics for all pods in the namespace
	podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	return marshal(podMetrics)
}
