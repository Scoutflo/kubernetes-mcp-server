package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getPrometheusURL attempts to discover the Prometheus URL in the cluster
func (k *Kubernetes) getPrometheusURL() (string, error) {
	// Try to find Prometheus Ingress first (preferred method)
	namespace := "scoutflo-monitoring"

	// Get Ingress resources in the monitoring namespace
	networkingV1 := k.clientSet.NetworkingV1()
	ingresses, err := networkingV1.Ingresses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err == nil && len(ingresses.Items) > 0 {
		// Look for Ingress that routes to Prometheus
		for _, ingress := range ingresses.Items {
			// Check all rules
			for _, rule := range ingress.Spec.Rules {
				// Check if the host looks like a monitoring host
				if strings.Contains(rule.Host, "scoutflo-monitoring") ||
					strings.Contains(rule.Host, "prometheus") {

					// Check all paths
					if rule.HTTP != nil {
						for _, path := range rule.HTTP.Paths {
							if strings.Contains(path.Backend.Service.Name, "prometheus") &&
								!strings.Contains(path.Backend.Service.Name, "alertmanager") {

								// Found a matching Ingress rule for Prometheus
								// Check if TLS is configured
								protocol := "http"
								for _, tls := range ingress.Spec.TLS {
									for _, host := range tls.Hosts {
										if host == rule.Host || host == "*.scoutflo.agency" {
											protocol = "https"
											break
										}
									}
								}

								// Construct URL with path if specified
								pathPrefix := "/"
								if path.Path != "" {
									pathPrefix = path.Path
								}

								prometheusURL := fmt.Sprintf("%s://%s%s", protocol, rule.Host, pathPrefix)
								return prometheusURL, nil
							}
						}
					}
				}
			}
		}
	}

	// If we couldn't find Prometheus via Ingress, try to port-forward the service
	// Note: This is a fallback and may need adaptation based on your environment
	return "", fmt.Errorf("prometheus URL not found via ingress")
}

// QueryPrometheus sends an instant query to Prometheus
func (k *Kubernetes) QueryPrometheus(query string) (string, error) {
	// Get Prometheus URL using service discovery
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the API query URL
	apiURL := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, url.QueryEscape(query))

	// Send the request
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to query Prometheus: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Prometheus response: %v", err)
	}

	// Convert result to JSON string
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Prometheus result: %v", err)
	}

	return string(resultJSON), nil
}

// QueryPrometheusRange sends a range query to Prometheus
func (k *Kubernetes) QueryPrometheusRange(query string, start, end time.Time, step string) (string, error) {
	// Get Prometheus URL using service discovery
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Convert times to Unix timestamps
	startTS := start.Unix()
	endTS := end.Unix()

	// Build the API query URL for range queries
	apiURL := fmt.Sprintf("%s/api/v1/query_range?query=%s&start=%d&end=%d&step=%s",
		prometheusURL, url.QueryEscape(query), startTS, endTS, url.QueryEscape(step))

	// Send the request
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to query Prometheus: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Prometheus response: %v", err)
	}

	// Convert result to JSON string
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Prometheus result: %v", err)
	}

	return string(resultJSON), nil
}

// ListPrometheusMetrics retrieves all available metrics from Prometheus
func (k *Kubernetes) ListPrometheusMetrics() (string, error) {
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the API URL for metadata endpoint
	apiURL := fmt.Sprintf("%s/api/v1/metadata", prometheusURL)

	// Send the request
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to list Prometheus metrics: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Prometheus response: %v", err)
	}

	// Convert result to JSON string
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Prometheus result: %v", err)
	}

	return string(resultJSON), nil
}

// GetPrometheusMetricInfo retrieves detailed information about a specific metric
func (k *Kubernetes) GetPrometheusMetricInfo(metricName string, includeStats bool) (string, error) {
	prometheusURL, err := k.getPrometheusURL()
	if err != nil {
		return "", fmt.Errorf("failed to discover Prometheus: %v", err)
	}

	// Build the API URL for metadata of specific metric
	metadataURL := fmt.Sprintf("%s/api/v1/metadata?metric=%s", prometheusURL, url.QueryEscape(metricName))

	// Send the request
	metadataResp, err := http.Get(metadataURL)
	if err != nil {
		return "", fmt.Errorf("failed to get Prometheus metric metadata: %v", err)
	}
	defer metadataResp.Body.Close()

	// Parse the metadata response
	var metadataResult map[string]interface{}
	if err := json.NewDecoder(metadataResp.Body).Decode(&metadataResult); err != nil {
		return "", fmt.Errorf("failed to decode Prometheus metadata response: %v", err)
	}

	result := map[string]interface{}{
		"name":     metricName,
		"metadata": metadataResult,
	}

	// If statistics are requested, fetch them
	if includeStats {
		statsResult, err := k.getMetricStatistics(prometheusURL, metricName)
		if err != nil {
			return "", err
		}
		result["statistics"] = statsResult
	}

	// Convert final result to JSON string
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Prometheus result: %v", err)
	}

	return string(resultJSON), nil
}

// getMetricStatistics calculates basic statistics for a metric
func (k *Kubernetes) getMetricStatistics(prometheusURL, metricName string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Define the queries for min, max, and count
	queries := map[string]string{
		"count": fmt.Sprintf("count(%s)", metricName),
		"min":   fmt.Sprintf("min(%s)", metricName),
		"max":   fmt.Sprintf("max(%s)", metricName),
		"avg":   fmt.Sprintf("avg(%s)", metricName),
	}

	// Execute each query and collect results
	for statName, query := range queries {
		queryURL := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, url.QueryEscape(query))
		resp, err := http.Get(queryURL)
		if err != nil {
			return nil, fmt.Errorf("failed to query %s statistic: %v", statName, err)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode %s statistic response: %v", statName, err)
		}
		resp.Body.Close()

		// Extract the value from Prometheus result
		if status, ok := result["status"].(string); ok && status == "success" {
			if data, ok := result["data"].(map[string]interface{}); ok {
				if resultData, ok := data["result"].([]interface{}); ok && len(resultData) > 0 {
					if resultItem, ok := resultData[0].(map[string]interface{}); ok {
						if value, ok := resultItem["value"].([]interface{}); ok && len(value) > 1 {
							stats[statName] = value[1]
						}
					}
				}
			}
		}
	}

	return stats, nil
}
