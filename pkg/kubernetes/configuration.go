package kubernetes

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"
)

func ConfigurationView(minify bool) (string, error) {
	var cfg clientcmdapi.Config
	var err error
	inClusterConfig, err := InClusterConfig()
	if err == nil && inClusterConfig != nil {
		cfg = *clientcmdapi.NewConfig()
		cfg.Clusters["cluster"] = &clientcmdapi.Cluster{
			Server:                inClusterConfig.Host,
			InsecureSkipTLSVerify: inClusterConfig.Insecure,
		}
		cfg.AuthInfos["user"] = &clientcmdapi.AuthInfo{
			Token: inClusterConfig.BearerToken,
		}
		cfg.Contexts["context"] = &clientcmdapi.Context{
			Cluster:  "cluster",
			AuthInfo: "user",
		}
		cfg.CurrentContext = "context"
	} else if cfg, err = resolveConfig().RawConfig(); err != nil {
		return "", err
	}
	if minify {
		if err = clientcmdapi.MinifyConfig(&cfg); err != nil {
			return "", err
		}
	}
	if err = clientcmdapi.FlattenConfig(&cfg); err != nil {
		// ignore error
		//return "", err
	}
	convertedObj, err := latest.Scheme.ConvertToVersion(&cfg, latest.ExternalVersion)
	if err != nil {
		return "", err
	}
	return marshal(convertedObj)
}

// GetAvailableAPIResources fetches all available API resources in the cluster
func GetAvailableAPIResources(ctx context.Context) (string, error) {
	// Create a new Kubernetes client
	k, err := NewKubernetes()
	if err != nil {
		return "", fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
	defer k.Close()

	// Get list of API groups
	groups, err := k.discoveryClient.ServerGroups()
	if err != nil {
		return "", fmt.Errorf("failed to get API groups: %v", err)
	}

	// Build a structured response
	groupVersions := []string{}
	apiResources := map[string][]map[string]interface{}{}

	// Process core API group (v1)
	resources, err := k.discoveryClient.ServerResourcesForGroupVersion("v1")
	if err == nil {
		groupVersions = append(groupVersions, "v1")
		apiResources["v1"] = formatAPIResources(resources.APIResources, "v1")
	}

	// Process other API groups
	for _, group := range groups.Groups {
		for _, version := range group.Versions {
			groupVersion := version.GroupVersion
			resources, err := k.discoveryClient.ServerResourcesForGroupVersion(groupVersion)
			if err != nil {
				continue
			}
			groupVersions = append(groupVersions, groupVersion)
			apiResources[groupVersion] = formatAPIResources(resources.APIResources, groupVersion)
		}
	}

	// Create the result structure
	result := map[string]interface{}{
		"apiGroups": groupVersions,
		"resources": apiResources,
	}

	// Marshal to YAML
	yamlResult, err := marshal(result)
	if err != nil {
		return "", err
	}

	return yamlResult, nil
}

// Helper function to format API resources
func formatAPIResources(resources []metav1.APIResource, groupVersion string) []map[string]interface{} {
	var formattedResources []map[string]interface{}

	for _, resource := range resources {
		// Skip subresources (those with '/')
		if strings.Contains(resource.Name, "/") {
			continue
		}

		// Format each resource
		formattedResource := map[string]interface{}{
			"name":       resource.Name,
			"kind":       resource.Kind,
			"namespaced": resource.Namespaced,
			"verbs":      resource.Verbs,
		}

		if resource.ShortNames != nil && len(resource.ShortNames) > 0 {
			formattedResource["shortNames"] = resource.ShortNames
		}

		formattedResources = append(formattedResources, formattedResource)
	}

	return formattedResources
}
