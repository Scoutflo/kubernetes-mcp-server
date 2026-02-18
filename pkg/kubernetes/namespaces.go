package kubernetes

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (k *Kubernetes) NamespacesList(ctx context.Context, limit int64, continueToken string) ([]byte, string, int64, error) {
	// Use slim response for MCP to reduce payload size
	return k.ResourcesListSlim(ctx, &schema.GroupVersionKind{
		Group: "", Version: "v1", Kind: "Namespace",
	}, "", limit, continueToken)
}
