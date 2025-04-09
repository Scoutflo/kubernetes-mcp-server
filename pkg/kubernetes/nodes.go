package kubernetes

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodesList returns a list of all nodes in the cluster
func (k *Kubernetes) NodesList(ctx context.Context) (string, error) {
	nodes, err := k.clientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	return marshal(nodes)
}

// NodesGet returns detailed information about a specific node
func (k *Kubernetes) NodesGet(ctx context.Context, name string) (string, error) {
	node, err := k.clientSet.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return marshal(node)
}
