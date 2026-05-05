package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

func (s *Server) initResources() []server.ServerTool {
	commonApiVersion := "v1 Pod, v1 Service, v1 Node, apps/v1 Deployment, networking.k8s.io/v1 Ingress"
	commonApiVersion = fmt.Sprintf("(common apiVersion and kind include: %s)", commonApiVersion)
	return []server.ServerTool{
		{Tool: WithMeta(
			mcp.NewTool("resources_list",
				mcp.WithDescription("List Kubernetes resources of any type. Use as a fallback when dedicated tools (pods_list, namespaces_list, nodes_list) are not available for the resource type. Requires exact apiVersion and kind — use get_available_API_resources to discover valid combinations. Provide namespace for namespaced resources to avoid cluster-wide scans."),
				mcp.WithString("apiVersion",
					mcp.Description("apiVersion of the resources (examples of valid apiVersion are: v1, apps/v1, networking.k8s.io/v1)"),
					mcp.Required(),
				),
				mcp.WithString("kind",
					mcp.Description("kind of the resources (examples of valid kind are: Pod, Service, Deployment, Ingress)"),
					mcp.Required(),
				),
				mcp.WithString("namespace",
					mcp.Description("Optional Namespace to retrieve the namespaced resources from (ignored in case of cluster scoped resources). If not provided, will list resources from all namespaces"),
				),
				mcp.WithNumber("limit", mcp.Description("Maximum number of items to return (default 10)")),
				mcp.WithString("continue", mcp.Description("Continuation token for pagination from a previous response")),
			),
			map[string]any{"provider": ProviderKubernetes},
		), Handler: s.resourcesList},
		{Tool: WithMeta(
			mcp.NewTool("resources_get",
				mcp.WithDescription("Retrieve a specific Kubernetes resource by name. Prefer dedicated tools (pods_get, nodes_get, helm_get_release) when available. Use for resource types without a dedicated tool. Requires exact apiVersion, kind, and name — use resources_list first to discover names when unknown."),
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
			),
			map[string]any{"provider": ProviderKubernetes},
		), Handler: s.resourcesGet},
		{Tool: WithMeta(
			mcp.NewTool("resources_create_or_update",
				mcp.WithDescription("Create or update a Kubernetes resource from a YAML or JSON manifest. Write operation. For partial updates to existing resources, prefer resources_patch which avoids full replacement. Ensure the manifest includes apiVersion, kind, metadata.name, and spec. Specify metadata.namespace for namespaced resources."),
				mcp.WithString("resource",
					mcp.Description("A JSON or YAML containing a representation of the Kubernetes resource. Should include top-level fields such as apiVersion,kind,metadata, and spec"),
					mcp.Required(),
				),
			),
			map[string]any{
				"provider": ProviderKubernetes,
				"hitl": map[string]any{
					"required":     true,
					"riskLevel":    RiskMedium,
					"approvalType": "single",
					"message":      "This will create or update a Kubernetes resource. Proceed?",
				},
			},
		), Handler: s.resourcesCreateOrUpdate},
		{Tool: WithMeta(
			mcp.NewTool("resources_delete",
				mcp.WithDescription("Delete a Kubernetes resource. Write operation — irreversible. Verify the resource exists with resources_get before calling. For pods managed by controllers, deletion triggers recreation; for other resource types, deletion is permanent."),
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
			),
			map[string]any{
				"provider": ProviderKubernetes,
				"hitl": map[string]any{
					"required":     true,
					"riskLevel":    RiskCritical,
					"approvalType": "single",
					"message":      "This will permanently delete a Kubernetes resource. This action cannot be undone. Proceed?",
				},
			},
		), Handler: s.resourcesDelete},
		{Tool: WithMeta(
			mcp.NewTool("get_resources_yaml",
				mcp.WithDescription("Retrieve Kubernetes resources in YAML format. Returns resource definitions as YAML for a specific resource or all resources of a given type. Use when you need YAML output for resources, export configurations, or work with YAML-based tools. Supports both single resource retrieval and listing. Requires apiVersion and kind. Optional name to get a specific resource, or omit to list all resources of the type. Common examples: v1 Pod, v1 Service, v1 Node, apps/v1 Deployment, networking.k8s.io/v1 Ingress"),
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
				mcp.WithNumber("limit", mcp.Description("Maximum number of items to return when listing (default 10). Only used when name is not provided")),
				mcp.WithString("continue", mcp.Description("Continuation token for pagination from a previous response. Only used when name is not provided")),
			),
			map[string]any{"provider": ProviderKubernetes},
		), Handler: s.resourcesYaml},
		// {Tool: mcp.NewTool("apply_manifest",
		// 	mcp.WithDescription("Apply a YAML resource file to the Kubernetes cluster"),
		// 	mcp.WithString("manifest_path",
		// 		mcp.Description("The path to the manifest file to apply (either this or yaml_content must be provided)"),
		// 	),
		// 	mcp.WithString("yaml_content",
		// 		mcp.Description("The raw YAML content to apply (either this or manifest_path must be provided)"),
		// 	),
		// ), Handler: s.applyManifest},
		{Tool: WithMeta(
			mcp.NewTool("resources_patch",
				mcp.WithDescription("Apply partial updates to a Kubernetes resource without replacing the entire resource. Supports three patch types: json (RFC 6902), merge (RFC 7396), and strategic (Kubernetes default). Returns the patched resource. Use when you need to modify specific fields of a resource efficiently. Requires apiVersion, kind, resource name, patch object, and optional namespace."),
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
			),
			map[string]any{
				"provider": ProviderKubernetes,
				"hitl": map[string]any{
					"required":     true,
					"riskLevel":    RiskMedium,
					"approvalType": "single",
					"message":      "This will patch and modify a Kubernetes resource. Proceed?",
				},
			},
		), Handler: s.resourcesPatch},
	}
}

func (s *Server) resourcesList(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: resources_list failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	namespace, err := ctr.RequireString("namespace")
	if err != nil {
		namespace = ""
	}

	limit := ctr.GetInt("limit", 10)

	continueToken := ctr.GetString("continue", "")
	view := ctr.GetString("view", "json")

	gvk, err := parseGroupVersionKind(ctr.GetRawArguments().(map[string]interface{}))
	if err != nil {
		klog.Errorf("Tool call: resources_list failed after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to list resources, %s", err)), nil
	}

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: resources_list - apiVersion: %s, kind: %s, namespace: %s, view: %s - got called by session id: %s", gvk.Version, gvk.Kind, namespace, view, sessionID)

	// Use slim response for MCP tools to reduce payload size
	ret, freshContinueToken, remainingCount, err := k.ResourcesListSlim(ctx, gvk, namespace, int64(limit), continueToken)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: resources_list failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to list resources: %v", err)), nil
	}

	var data interface{}
	if err := json.Unmarshal(ret, &data); err != nil {
		klog.Errorf("Tool call: resources_list failed to unmarshal response after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to unmarshal resource list: %v", err)), nil
	}

	response := ListResourceToolOutput{
		Data:                data,
		ContinueToken:       freshContinueToken,
		RemainingItemsCount: remainingCount,
	}

	if view == "yaml" {
		yamlBytes, err := yaml.Marshal(response)
		if err != nil {
			klog.Errorf("Tool call: resources_list failed to marshal result to YAML after %v: %v", duration, err)
			return NewTextResult("", fmt.Errorf("failed to marshal resource list to YAML: %v", err)), nil
		}
		klog.V(1).Infof("Tool call: resources_list completed successfully in %v by session id: %s", duration, sessionID)
		return NewTextResult(string(yamlBytes), nil), nil
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		klog.Errorf("Tool call: resources_list failed to marshal result to JSON after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to marshal resource list: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: resources_list completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(string(jsonBytes), nil), nil
}

func (s *Server) resourcesGet(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: resources_get failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
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

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: resources_get - apiVersion: %s, kind: %s, namespace: %s, name: %s - got called by session id: %s", gvk.Version, gvk.Kind, namespace, name, sessionID)

	ret, err := k.ResourcesGet(ctx, gvk, namespace, name)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: resources_get failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to get resource: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: resources_get completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, err), nil
}

func (s *Server) resourcesCreateOrUpdate(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: resources_create_or_update failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
	resource, err := ctr.RequireString("resource")
	if err != nil {
		klog.Errorf("Tool call: resources_create_or_update failed after %v: missing resource parameter", time.Since(start))
		return NewTextResult("", errors.New("failed to create or update resources, missing argument resource")), nil
	}

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: resources_create_or_update - resource YAML/JSON provided (length: %d) - got called by session id: %s", len(resource), sessionID)

	ret, err := k.ResourcesCreateOrUpdate(ctx, resource)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: resources_create_or_update failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to create or update resources: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: resources_create_or_update completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, err), nil
}

func (s *Server) resourcesDelete(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: resources_delete failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
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

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: resources_delete - apiVersion: %s, kind: %s, namespace: %s, name: %s - got called by session id: %s", gvk.Version, gvk.Kind, namespace, name, sessionID)

	err = k.ResourcesDelete(ctx, gvk, namespace, name)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: resources_delete failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to delete resource: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: resources_delete completed successfully in %v by session id: %s", duration, sessionID)
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

// func (s *Server) applyManifest(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
// 	start := time.Now()
// 	k, err := s.getKubernetesClient(ctr)
// 	if err != nil {
// 		klog.Errorf("Tool call: apply_manifest failed to get Kubernetes client after %v: %v", time.Since(start), err)
// 		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
// 	}
// 	manifestPath := ctr.GetString("manifest_path", "")
// 	yamlContent := ctr.GetString("yaml_content", "")

// 	sessionID := getSessionID(ctx)
// 	klog.V(1).Infof("Tool: apply_manifest - manifest_path: %s, yaml_content_length: %d - got called by session id: %s", manifestPath, len(yamlContent), sessionID)

// 	// Ensure at least one of manifest_path or yaml_content is provided
// 	if manifestPath == "" && yamlContent == "" {
// 		klog.Errorf("Tool call: apply_manifest failed after %v: neither manifest_path nor yaml_content provided", time.Since(start))
// 		return NewTextResult("", errors.New("failed to apply manifest, either manifest_path or yaml_content must be provided")), nil
// 	}

// 	var content string

// 	// If manifest_path is provided, read the file
// 	if manifestPath != "" {
// 		contentBytes, err := os.ReadFile(manifestPath)
// 		if err != nil {
// 			klog.Errorf("Tool call: apply_manifest failed after %v: failed to read file %s: %v", time.Since(start), manifestPath, err)
// 			return NewTextResult("", fmt.Errorf("failed to read manifest file: %v", err)), nil
// 		}
// 		content = string(contentBytes)
// 	} else {
// 		// Otherwise use the provided yaml_content
// 		content = yamlContent
// 	}

// 	// Apply the manifest content
// 	ret, err := k.ResourcesCreateOrUpdate(ctx, content)
// 	duration := time.Since(start)

// 	if err != nil {
// 		klog.Errorf("Tool call: apply_manifest failed after %v: %v", duration, err)
// 		return NewTextResult("", fmt.Errorf("failed to apply manifest: %v", err)), nil
// 	}

// 	klog.V(1).Infof("Tool call: apply_manifest completed successfully in %v by session id: %s", duration, sessionID)
// 	return NewTextResult(ret, nil), nil
// }

func (s *Server) resourcesPatch(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: resources_patch failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}
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

	// Get patch as an object from raw arguments
	rawArgs := ctr.GetRawArguments().(map[string]interface{})
	patchObj, exists := rawArgs["patch"]
	if !exists {
		klog.Errorf("Tool call: resources_patch failed after %v: missing patch parameter", time.Since(start))
		return NewTextResult("", errors.New("failed to patch resource, missing argument patch")), nil
	}

	patchType := "strategic"
	if pt, err := ctr.RequireString("patch_type"); err == nil {
		patchType = pt
	}

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: resources_patch - apiVersion: %s, kind: %s, namespace: %s, name: %s, patch_type: %s - got called by session id: %s", gvk.Version, gvk.Kind, namespace, resourceName, patchType, sessionID)

	// Validate patch type
	if patchType != "json" && patchType != "merge" && patchType != "strategic" {
		klog.Errorf("Tool call: resources_patch failed after %v: invalid patch_type: %s", time.Since(start), patchType)
		return NewTextResult("", fmt.Errorf("invalid patch_type: %s. Must be one of: json, merge, strategic", patchType)), nil
	}

	// Convert the patch object to JSON
	patchJSON, err := json.Marshal(patchObj)
	if err != nil {
		klog.Errorf("Tool call: resources_patch failed after %v: failed to marshal patch: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to marshal patch data: %v", err)), nil
	}

	// Apply the patch
	ret, err := k.ResourcesPatch(ctx, gvk, namespace, resourceName, patchType, patchJSON)
	duration := time.Since(start)

	if err != nil {
		klog.Errorf("Tool call: resources_patch failed after %v: %v", duration, err)
		return NewTextResult("", fmt.Errorf("failed to patch resource: %v", err)), nil
	}

	klog.V(1).Infof("Tool call: resources_patch completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(ret, nil), nil
}

func (s *Server) resourcesYaml(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	k, err := s.getKubernetesClient(ctr)
	if err != nil {
		klog.Errorf("Tool call: get_resources_yaml failed to get Kubernetes client after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to initialize Kubernetes client: %v", err)), nil
	}

	namespace, err := ctr.RequireString("namespace")
	if err != nil {
		namespace = ""
	}

	name := ctr.GetString("name", "")

	gvk, err := parseGroupVersionKind(ctr.GetRawArguments().(map[string]interface{}))
	if err != nil {
		klog.Errorf("Tool call: get_resources_yaml failed after %v: %v", time.Since(start), err)
		return NewTextResult("", fmt.Errorf("failed to get resource YAML, %s", err)), nil
	}

	sessionID := getSessionID(ctx)
	klog.V(1).Infof("Tool: get_resources_yaml - apiVersion: %s, kind: %s, namespace: %s, name: %s - got called by session id: %s", gvk.Version, gvk.Kind, namespace, name, sessionID)

	var yamlOutput string

	if name != "" {
		ret, err := k.ResourcesGet(ctx, gvk, namespace, name)
		duration := time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: get_resources_yaml failed after %v: %v", duration, err)
			return NewTextResult("", fmt.Errorf("failed to get resource YAML: %v", err)), nil
		}

		var resourceData interface{}
		if err := json.Unmarshal([]byte(ret), &resourceData); err != nil {
			klog.Errorf("Tool call: get_resources_yaml failed to unmarshal response after %v: %v", duration, err)
			return NewTextResult("", fmt.Errorf("failed to unmarshal resource: %v", err)), nil
		}

		yamlBytes, err := yaml.Marshal(resourceData)
		if err != nil {
			klog.Errorf("Tool call: get_resources_yaml failed to marshal result to YAML after %v: %v", duration, err)
			return NewTextResult("", fmt.Errorf("failed to marshal resource to YAML: %v", err)), nil
		}

		yamlOutput = string(yamlBytes)
	} else {
		limit := ctr.GetInt("limit", 10)
		continueToken := ctr.GetString("continue", "")

		// Use slim response for MCP tools to reduce payload size
		ret, freshContinueToken, remainingCount, err := k.ResourcesListSlim(ctx, gvk, namespace, int64(limit), continueToken)
		duration := time.Since(start)

		if err != nil {
			klog.Errorf("Tool call: get_resources_yaml failed after %v: %v", duration, err)
			return NewTextResult("", fmt.Errorf("failed to list resources: %v", err)), nil
		}

		var data interface{}
		if err := json.Unmarshal(ret, &data); err != nil {
			klog.Errorf("Tool call: get_resources_yaml failed to unmarshal response after %v: %v", duration, err)
			return NewTextResult("", fmt.Errorf("failed to unmarshal resource list: %v", err)), nil
		}

		response := ListResourceToolOutput{
			Data:                data,
			ContinueToken:       freshContinueToken,
			RemainingItemsCount: remainingCount,
		}

		yamlBytes, err := yaml.Marshal(response)
		if err != nil {
			klog.Errorf("Tool call: get_resources_yaml failed to marshal result to YAML after %v: %v", duration, err)
			return NewTextResult("", fmt.Errorf("failed to marshal resource list to YAML: %v", err)), nil
		}

		yamlOutput = string(yamlBytes)
	}

	duration := time.Since(start)
	klog.V(1).Infof("Tool call: get_resources_yaml completed successfully in %v by session id: %s", duration, sessionID)
	return NewTextResult(yamlOutput, nil), nil
}
