package kubernetes

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (k *Kubernetes) NamespacesList(ctx context.Context) (string, error) {
	return k.ResourcesList(ctx, &schema.GroupVersionKind{
		Group: "", Version: "v1", Kind: "Namespace",
	}, "")
}
