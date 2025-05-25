package mcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initLabels() []server.ServerTool {
	commonApiVersion := "v1 Pod, v1 Service, v1 Node, apps/v1 Deployment, networking.k8s.io/v1 Ingress"
	commonApiVersion = fmt.Sprintf("(common apiVersion and kind include: %s)", commonApiVersion)
	return []server.ServerTool{
		{Tool: mcp.NewTool("label_resource",
			mcp.WithDescription("Apply labels to a Kubernetes resource\n"+
				commonApiVersion),
			mcp.WithString("apiVersion",
				mcp.Description("apiVersion of the resource (examples of valid apiVersion are: v1, apps/v1, networking.k8s.io/v1)"),
				mcp.Required(),
			),
			mcp.WithString("kind",
				mcp.Description("kind of the resource (examples of valid kind are: Pod, Service, Deployment, Ingress)"),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("Name of the resource"),
				mcp.Required(),
			),
			mcp.WithString("namespace",
				mcp.Description("Namespace where the resource is located (ignored for cluster-scoped resources)"),
			),
			mcp.WithObject("labels",
				mcp.Description("The labels to apply to the resource as key-value pairs"),
				mcp.Required(),
			),
		), Handler: s.labelResource},

		{Tool: mcp.NewTool("remove_label",
			mcp.WithDescription("Remove a label from a Kubernetes resource\n"+
				commonApiVersion),
			mcp.WithString("apiVersion",
				mcp.Description("apiVersion of the resource (examples of valid apiVersion are: v1, apps/v1, networking.k8s.io/v1)"),
				mcp.Required(),
			),
			mcp.WithString("kind",
				mcp.Description("kind of the resource (examples of valid kind are: Pod, Service, Deployment, Ingress)"),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("Name of the resource"),
				mcp.Required(),
			),
			mcp.WithString("namespace",
				mcp.Description("Namespace where the resource is located (ignored for cluster-scoped resources)"),
			),
			mcp.WithString("label_key",
				mcp.Description("The key of the label to remove"),
				mcp.Required(),
			),
		), Handler: s.removeLabel},

		{Tool: mcp.NewTool("annotate_resource",
			mcp.WithDescription("Apply annotations to a Kubernetes resource\n"+
				commonApiVersion),
			mcp.WithString("apiVersion",
				mcp.Description("apiVersion of the resource (examples of valid apiVersion are: v1, apps/v1, networking.k8s.io/v1)"),
				mcp.Required(),
			),
			mcp.WithString("kind",
				mcp.Description("kind of the resource (examples of valid kind are: Pod, Service, Deployment, Ingress)"),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("Name of the resource"),
				mcp.Required(),
			),
			mcp.WithString("namespace",
				mcp.Description("Namespace where the resource is located (ignored for cluster-scoped resources)"),
			),
			mcp.WithObject("annotations",
				mcp.Description("The annotations to apply to the resource as key-value pairs"),
				mcp.Required(),
			),
		), Handler: s.annotateResource},

		{Tool: mcp.NewTool("remove_annotation",
			mcp.WithDescription("Remove an annotation from a Kubernetes resource\n"+
				commonApiVersion),
			mcp.WithString("apiVersion",
				mcp.Description("apiVersion of the resource (examples of valid apiVersion are: v1, apps/v1, networking.k8s.io/v1)"),
				mcp.Required(),
			),
			mcp.WithString("kind",
				mcp.Description("kind of the resource (examples of valid kind are: Pod, Service, Deployment, Ingress)"),
				mcp.Required(),
			),
			mcp.WithString("name",
				mcp.Description("Name of the resource"),
				mcp.Required(),
			),
			mcp.WithString("namespace",
				mcp.Description("Namespace where the resource is located (ignored for cluster-scoped resources)"),
			),
			mcp.WithString("annotation_key",
				mcp.Description("The key of the annotation to remove"),
				mcp.Required(),
			),
		), Handler: s.removeAnnotation},
	}
}

func (s *Server) labelResource(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := ctr.GetString("namespace", "")

	gvk, err := parseGroupVersionKind(ctr.GetRawArguments().(map[string]interface{}))
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to label resource, %s", err)), nil
	}

	name := ctr.GetString("name", "")
	if name == "" {
		return NewTextResult("", errors.New("failed to label resource, missing argument name")), nil
	}

	args := ctr.GetRawArguments()
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return NewTextResult("", errors.New("failed to get arguments")), nil
	}

	labels, ok := argsMap["labels"].(map[string]interface{})
	if !ok || len(labels) == 0 {
		return NewTextResult("", errors.New("failed to label resource, missing or invalid labels")), nil
	}

	// Convert labels to string map
	labelMap := make(map[string]string)
	for k, v := range labels {
		labelMap[k] = fmt.Sprintf("%v", v)
	}

	ret, err := s.k.LabelResource(ctx, gvk, namespace, name, labelMap)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to label resource: %v", err)), nil
	}

	return NewTextResult(ret, nil), nil
}

func (s *Server) removeLabel(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := ctr.GetString("namespace", "")

	gvk, err := parseGroupVersionKind(ctr.GetRawArguments().(map[string]interface{}))
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to remove label, %s", err)), nil
	}

	name := ctr.GetString("name", "")
	if name == "" {
		return NewTextResult("", errors.New("failed to remove label, missing argument name")), nil
	}

	labelKey := ctr.GetString("label_key", "")
	if labelKey == "" {
		return NewTextResult("", errors.New("failed to remove label, missing or invalid label_key")), nil
	}

	ret, err := s.k.RemoveLabel(ctx, gvk, namespace, name, labelKey)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to remove label: %v", err)), nil
	}

	return NewTextResult(ret, nil), nil
}

func (s *Server) annotateResource(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := ctr.GetString("namespace", "")

	gvk, err := parseGroupVersionKind(ctr.GetRawArguments().(map[string]interface{}))
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to annotate resource, %s", err)), nil
	}

	name := ctr.GetString("name", "")
	if name == "" {
		return NewTextResult("", errors.New("failed to annotate resource, missing argument name")), nil
	}

	args := ctr.GetRawArguments()
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return NewTextResult("", errors.New("failed to get arguments")), nil
	}

	annotations, ok := argsMap["annotations"].(map[string]interface{})
	if !ok || len(annotations) == 0 {
		return NewTextResult("", errors.New("failed to annotate resource, missing or invalid annotations")), nil
	}

	// Convert annotations to string map
	annotationMap := make(map[string]string)
	for k, v := range annotations {
		annotationMap[k] = fmt.Sprintf("%v", v)
	}

	ret, err := s.k.AnnotateResource(ctx, gvk, namespace, name, annotationMap)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to annotate resource: %v", err)), nil
	}

	return NewTextResult(ret, nil), nil
}

func (s *Server) removeAnnotation(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := ctr.GetString("namespace", "")

	gvk, err := parseGroupVersionKind(ctr.GetRawArguments().(map[string]interface{}))
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to remove annotation, %s", err)), nil
	}

	name := ctr.GetString("name", "")
	if name == "" {
		return NewTextResult("", errors.New("failed to remove annotation, missing argument name")), nil
	}

	annotationKey := ctr.GetString("annotation_key", "")
	if annotationKey == "" {
		return NewTextResult("", errors.New("failed to remove annotation, missing or invalid annotation_key")), nil
	}

	ret, err := s.k.RemoveAnnotation(ctx, gvk, namespace, name, annotationKey)
	if err != nil {
		return NewTextResult("", fmt.Errorf("failed to remove annotation: %v", err)), nil
	}

	return NewTextResult(ret, nil), nil
}
