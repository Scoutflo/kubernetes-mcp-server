package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

func (s *Server) initResources() []server.ServerTool {
	commonApiVersion := "v1 Pod, v1 Service, v1 Node, apps/v1 Deployment, networking.k8s.io/v1 Ingress"
	commonApiVersion = fmt.Sprintf("(common apiVersion and kind include: %s)", commonApiVersion)
	return []server.ServerTool{
		{Tool: mcp.NewTool("resources_list",
			mcp.WithDescription("List Kubernetes resources and objects in the current cluster by providing their apiVersion and kind and optionally the namespace\n"+
				commonApiVersion),
			mcp.WithString("apiVersion",
				mcp.Description("apiVersion of the resources (examples of valid apiVersion are: v1, apps/v1, networking.k8s.io/v1)"),
				mcp.Required(),
			),
			mcp.WithString("kind",
				mcp.Description("kind of the resources (examples of valid kind are: Pod, Service, Deployment, Ingress)"),
				mcp.Required(),
			),
			mcp.WithString("namespace",
				mcp.Description("Optional Namespace to retrieve the namespaced resources from (ignored in case of cluster scoped resources). If not provided, will list resources from all namespaces"))),
			Handler: s.resourcesList,
		},
		{Tool: mcp.NewTool("resources_get",
			mcp.WithDescription("Get a Kubernetes resource in the current cluster by providing its apiVersion, kind, optionally the namespace, and its name\n"+
				commonApiVersion),
			mcp.WithString("apiVersion",
				mcp.Description("apiVersion of the resource (examples of valid apiVersion are: v1, apps/v1, networking.k8s.io/v1)"),
				mcp.Required(),
			),
			mcp.WithString("kind",
				mcp.Description("kind of the resource (examples of valid kind are: Pod, Service, Deployment, Ingress)"),
				mcp.Required(),
			),
			mcp.WithString("namespace",
				mcp.Description("Optional Namespace to retrieve the namespaced resource from (ignored in case of cluster scoped resources). If not provided, will get resource from configured namespace"),
			),
			mcp.WithString("name", mcp.Description("Name of the resource"), mcp.Required()),
		), Handler: s.resourcesGet},
		{Tool: mcp.NewTool("resources_create_or_update",
			mcp.WithDescription("Create or update a Kubernetes resource in the current cluster by providing a YAML or JSON representation of the resource\n"+
				commonApiVersion),
			mcp.WithString("resource",
				mcp.Description("A JSON or YAML containing a representation of the Kubernetes resource. Should include top-level fields such as apiVersion,kind,metadata, and spec"),
				mcp.Required(),
			),
		), Handler: s.resourcesCreateOrUpdate},
		{Tool: mcp.NewTool("resources_delete",
			mcp.WithDescription("Delete a Kubernetes resource in the current cluster by providing its apiVersion, kind, optionally the namespace, and its name\n"+
				commonApiVersion),
			mcp.WithString("apiVersion",
				mcp.Description("apiVersion of the resource (examples of valid apiVersion are: v1, apps/v1, networking.k8s.io/v1)"),
				mcp.Required(),
			),
			mcp.WithString("kind",
				mcp.Description("kind of the resource (examples of valid kind are: Pod, Service, Deployment, Ingress)"),
				mcp.Required(),
			),
			mcp.WithString("namespace",
				mcp.Description("Optional Namespace to delete the namespaced resource from (ignored in case of cluster scoped resources). If not provided, will delete resource from configured namespace"),
			),
			mcp.WithString("name", mcp.Description("Name of the resource"), mcp.Required()),
		), Handler: s.resourcesDelete},
		{Tool: mcp.NewTool("get_resources_yaml",
			mcp.WithDescription("Get the YAML representation of a resource in Kubernetes\n"+
				commonApiVersion),
			mcp.WithString("apiVersion",
				mcp.Description("apiVersion of the resource (examples of valid apiVersion are: v1, apps/v1, networking.k8s.io/v1)"),
				mcp.Required(),
			),
			mcp.WithString("kind",
				mcp.Description("kind of the resource (examples of valid kind are: Pod, Service, Deployment, Ingress)"),
				mcp.Required(),
			),
			mcp.WithString("namespace",
				mcp.Description("The namespace of the resource to get the definition for"),
			),
			mcp.WithString("name", mcp.Description("The name of the resource to get the YAML definition for. If not provided, all resources of the given type will be returned")),
		), Handler: s.resourcesYaml},
		{Tool: mcp.NewTool("apply_manifest",
			mcp.WithDescription("Apply a YAML resource file to the Kubernetes cluster"),
			mcp.WithString("manifest_path",
				mcp.Description("The path to the manifest file to apply (either this or yaml_content must be provided)"),
			),
			mcp.WithString("yaml_content",
				mcp.Description("The raw YAML content to apply (either this or manifest_path must be provided)"),
			),
		), Handler: s.applyManifest},
		{Tool: mcp.NewTool("resources_patch",
			mcp.WithDescription("Patch a resource in Kubernetes\n"+
				commonApiVersion),
			mcp.WithString("apiVersion",
				mcp.Description("apiVersion of the resource (examples of valid apiVersion are: v1, apps/v1, networking.k8s.io/v1)"),
				mcp.Required(),
			),
			mcp.WithString("kind",
				mcp.Description("kind of the resource (examples of valid kind are: Pod, Service, Deployment, Ingress)"),
				mcp.Required(),
			),
			mcp.WithString("resource_name",
				mcp.Description("The name of the resource to patch"),
				mcp.Required(),
			),
			mcp.WithString("namespace",
				mcp.Description("The namespace of the resource to patch (ignored for cluster-scoped resources)"),
			),
			mcp.WithObject("patch",
				mcp.Description("The patch to apply to the resource as a JSON object"),
				mcp.Required(),
			),
			mcp.WithString("patch_type",
				mcp.Description("The type of patch to apply (json, merge, strategic). Defaults to strategic for Kubernetes resources"),
			),
		), Handler: s.resourcesPatch},
	}
}

func (s *Server) resourcesList(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	namespace, err := ctr.RequireString("namespace")
	if err != nil {
		namespace = ""
	}
	gvk, err := parseGroupVersionKind(ctr.GetRawArguments().(map[string]interface{}))
	if err != nil {
		klog.Errorf("Tool call: resources_list failed after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to list resources, %s", err)), nil
	}

	klog.V(1).Infof("Tool: resources_list - apiVersion: %s, kind: %s, namespace: %s - got called", gvk.Version, gvk.Kind, namespace)

	ret, err := s.k.ResourcesList(ctx, gvk, namespace)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: resources_list failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to list resources: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: resources_list completed successfully in %v", duration)
	return NewTextResult(ret, err), nil
}

func (s *Server) resourcesGet(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	namespace, err := ctr.RequireString("namespace")
	if err != nil {
		namespace = ""
	}
	gvk, err := parseGroupVersionKind(ctr.GetRawArguments().(map[string]interface{}))
	if err != nil {
		klog.Errorf("Tool call: resources_get failed after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to get resource, %s", err)), nil
	}
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: resources_get failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", errors.New("failed to get resource, missing argument name")), nil
	}

	klog.V(1).Infof("Tool: resources_get - apiVersion: %s, kind: %s, namespace: %s, name: %s - got called", gvk.Version, gvk.Kind, namespace, name)

	ret, err := s.k.ResourcesGet(ctx, gvk, namespace, name)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: resources_get failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get resource: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: resources_get completed successfully in %v", duration)
	return NewTextResult(ret, err), nil
}

func (s *Server) resourcesCreateOrUpdate(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	resource, err := ctr.RequireString("resource")
	if err != nil {
		klog.Errorf("Tool call: resources_create_or_update failed after %v: missing resource parameter", time.Since(start))
		return NewTextResult("", errors.New("failed to create or update resources, missing argument resource")), nil
	}

	klog.V(1).Infof("Tool: resources_create_or_update - resource YAML/JSON provided (length: %d) - got called", len(resource))

	ret, err := s.k.ResourcesCreateOrUpdate(ctx, resource)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: resources_create_or_update failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to create or update resources: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: resources_create_or_update completed successfully in %v", duration)
	return NewTextResult(ret, err), nil
}

func (s *Server) resourcesDelete(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	namespace, err := ctr.RequireString("namespace")
	if err != nil {
		namespace = ""
	}
	gvk, err := parseGroupVersionKind(ctr.GetRawArguments().(map[string]interface{}))
	if err != nil {
		klog.Errorf("Tool call: resources_delete failed after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to delete resource, %s", err)), nil
	}
	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: resources_delete failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", errors.New("failed to delete resource, missing argument name")), nil
	}

	klog.V(1).Infof("Tool: resources_delete - apiVersion: %s, kind: %s, namespace: %s, name: %s - got called", gvk.Version, gvk.Kind, namespace, name)

	err = s.k.ResourcesDelete(ctx, gvk, namespace, name)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: resources_delete failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to delete resource: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: resources_delete completed successfully in %v", duration)
	return NewTextResult("Resource deleted successfully", err), nil
}

func parseGroupVersionKind(arguments map[string]interface{}) (*schema.GroupVersionKind, error) {
	apiVersion := arguments["apiVersion"]
	if apiVersion == nil {
		return nil, errors.New("missing argument apiVersion")
	}
	kind := arguments["kind"]
	if kind == nil {
		return nil, errors.New("missing argument kind")
	}
	gv, err := schema.ParseGroupVersion(apiVersion.(string))
	if err != nil {
		return nil, errors.New("invalid argument apiVersion")
	}
	return &schema.GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: kind.(string)}, nil
}

func (s *Server) resourcesYaml(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	namespace, err := ctr.RequireString("namespace")
	if err != nil {
		namespace = ""
	}
	gvk, err := parseGroupVersionKind(ctr.GetRawArguments().(map[string]interface{}))
	if err != nil {
		klog.Errorf("Tool call: get_resources_yaml failed after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to get YAML, %s", err)), nil
	}

	name, err := ctr.RequireString("name")
	if err != nil {
		klog.Errorf("Tool call: get_resources_yaml failed after %v: missing name parameter", time.Since(start))
		return NewTextResult("", errors.New("failed to get YAML, missing argument name")), nil
	}

	klog.V(1).Infof("Tool: get_resources_yaml - apiVersion: %s, kind: %s, namespace: %s, name: %s - got called", gvk.Version, gvk.Kind, namespace, name)

	if name != "" {
		// Get a specific resource
		ret, err := s.k.ResourcesGet(ctx, gvk, namespace, name)
		duration := time.Since(start)
		if err != nil {
			klog.Errorf("Tool call: get_resources_yaml failed after %v: %v", duration, err)
			return NewTextResult("", fmt.Errorf("failed to get resource YAML: %v", err)), nil
		}
		klog.V(1).Infof("Tool call: get_resources_yaml completed successfully in %v", duration)
		return NewTextResult(ret, err), nil
	} else {
		// Get all resources of this type in the namespace
		ret, err := s.k.ResourcesList(ctx, gvk, namespace)
		duration := time.Since(start)
		if err != nil {
			klog.Errorf("Tool call: get_resources_yaml failed after %v: %v", duration, err)
			return NewTextResult("", fmt.Errorf("failed to list resources YAML: %v", err)), nil
		}
		klog.V(1).Infof("Tool call: get_resources_yaml completed successfully in %v", duration)
		return NewTextResult(ret, err), nil
	}
}

func (s *Server) applyManifest(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	manifestPath := ctr.GetString("manifest_path", "")
	yamlContent := ctr.GetString("yaml_content", "")

	klog.V(1).Infof("Tool: apply_manifest - manifest_path: %s, yaml_content_length: %d - got called", manifestPath, len(yamlContent))

	// Ensure at least one of manifest_path or yaml_content is provided
	if manifestPath == "" && yamlContent == "" {
		klog.Errorf("Tool call: apply_manifest failed after %v: neither manifest_path nor yaml_content provided", time.Since(start))
		return NewTextResult("", errors.New("failed to apply manifest, either manifest_path or yaml_content must be provided")), nil
	}

	var content string
	var err error

	// If manifest_path is provided, read the file
	if manifestPath != "" {
		contentBytes, err := os.ReadFile(manifestPath)
		if err != nil {
			klog.Errorf("Tool call: apply_manifest failed after %v: failed to read file %s: %v", time.Since(start), manifestPath, err)
			return NewTextResult("", fmt.Errorf("failed to read manifest file: %v", err)), nil
		}
		content = string(contentBytes)
	} else {
		// Otherwise use the provided yaml_content
		content = yamlContent
	}

	// Apply the manifest content
	ret, err := s.k.ResourcesCreateOrUpdate(ctx, content)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: apply_manifest failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to apply manifest: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: apply_manifest completed successfully in %v", duration)
	return NewTextResult(ret, nil), nil
}

func (s *Server) resourcesPatch(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	namespace, err := ctr.RequireString("namespace")
	if err != nil {
		namespace = ""
	}

	gvk, err := parseGroupVersionKind(ctr.GetRawArguments().(map[string]interface{}))
	if err != nil {
		klog.Errorf("Tool call: resources_patch failed after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to patch resource, %s", err)), nil
	}

	resourceName, err := ctr.RequireString("resource_name")
	if err != nil {
		klog.Errorf("Tool call: resources_patch failed after %v: missing resource_name parameter", time.Since(start))
		return NewTextResult("", errors.New("failed to patch resource, missing argument resource_name")), nil
	}

	patch, err := ctr.RequireString("patch")
	if err != nil {
		klog.Errorf("Tool call: resources_patch failed after %v: missing patch parameter", time.Since(start))
		return NewTextResult("", errors.New("failed to patch resource, missing argument patch")), nil
	}

	patchType := "strategic"
	if pt, err := ctr.RequireString("patch_type"); err == nil {
		patchType = pt
	}

	klog.V(1).Infof("Tool: resources_patch - apiVersion: %s, kind: %s, namespace: %s, name: %s, patch_type: %s - got called", gvk.Version, gvk.Kind, namespace, resourceName, patchType)

	// Validate patch type
	if patchType != "json" && patchType != "merge" && patchType != "strategic" {
		klog.Errorf("Tool call: resources_patch failed after %v: invalid patch_type: %s", time.Since(start), patchType)
		return NewTextResult("", fmt.Errorf("invalid patch_type: %s. Must be one of: json, merge, strategic", patchType)), nil
	}

	// Convert the patch to JSON
	patchJSON, err := json.Marshal(patch)
	if err != nil {
		klog.Errorf("Tool call: resources_patch failed after %v: failed to marshal patch: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to marshal patch data: %v", err)), nil
	}

	// Apply the patch
	ret, err := s.k.ResourcesPatch(ctx, gvk, namespace, resourceName, patchType, patchJSON)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: resources_patch failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to patch resource: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: resources_patch completed successfully in %v", duration)
	return NewTextResult(ret, nil), nil
}
