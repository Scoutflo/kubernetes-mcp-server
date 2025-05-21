package llm

// PromptGeneratorKnowledgeBase contains the knowledge base for generating structured prompts
const PromptGeneratorKnowledgeBase = `
# Prompt Generator Knowledge Base

Your task is to generate a well-structured, comprehensive prompt for analyzing Kubernetes clusters based on the user's description. The prompt should include parameters, step-by-step instructions, recommended tools, and a clear structure for presenting results.

## Format Instructions

The prompt should be structured in the following format:

1. Title - Brief description of the task
2. Parameters - List of parameters with placeholders using the format ${parameter_name}
3. Step-by-step instructions - Clear, numbered or bulleted steps
4. Tools to use - List of relevant Kubernetes tools from the available tools list
5. Any specific metrics or requirements that should be included
6. Root cause - Instructions for identifying and presenting root causes of issues add if related logs are provided (small section)
7. App or infra problem - Framework for determining if issues are application or infrastructure related (small section)
8. Evidence sections - Format for presenting supporting evidence for findings and recommendations (small section)

## Available Tools
Below is the list of available Kubernetes tools that can be included in your prompt:

### Cluster Information Tools
- configuration_view (Get Kubernetes configuration)
- get_available_API_resources (Get all available API resources)
- get_resources_yaml (Get YAML representation of resources)
- namespaces_list (List all namespaces)
- events_list (List Kubernetes events)

### Node Tools
- nodes_get (Get detailed node information)
- nodes_list (List all nodes)
- nodes_metrics (Get node metrics)

### Pod Tools
- pods_get (Get pod details)
- pods_list (List all pods)
- pods_list_in_namespace (List pods in namespace)
- pods_log (Get pod logs)
- pods_delete (Delete a pod)
- pods_exec (Execute commands in pod)
- pods_run (Run a pod)
- pods_metrics (Get pod metrics)

### Resource Management
- resources_create_or_update (Create/update resources)
- resources_delete (Delete resources)
- resources_get (Get specific resource)
- resources_list (List resources)
- resources_patch (Patch resources)
- apply_manifest (Apply YAML manifest)
- rollout (Manage deployments rollout)

### Networking Tools
- check_service_connectivity (Check service connectivity)
- check_ingress_connectivity (Check ingress connectivity)
- port_forward tools (create_port_forward, list_port_forward, cancel_port_forward)

### Label and Annotation Tools
- annotate_resource (Apply annotations)
- remove_annotation (Remove annotation)
- label_resource (Apply labels)
- remove_label (Remove label)

### Prometheus Monitoring Tools
- prometheus_generate_query (Generate PromQL queries)
- prometheus_metrics_query (Execute instant queries)
- prometheus_metrics_query_range (Execute range queries)
- prometheus_list_metrics (List available metrics)
- prometheus_get_alerts (Get active alerts)
- prometheus_get_rules (Get alerting rules)
- prometheus_create_alert (Create alert rule)
- prometheus_update_alert (Update alert rule)
- prometheus_delete_alert (Delete alert rule)

### Argo CD Tools
- argocd_list_applications (List ArgoCD applications)
- argocd_get_application (Get application details)
- argocd_sync_application (Sync application)
- argocd_create_application (Create application)
- argocd_update_application (Update application)
- argocd_delete_application (Delete application)

### Argo Rollout Tools
- create_argo_rollout_config (Generate rollout config)
- get_argo_rollout (Get rollout status)
- promote_argo_rollout (Promote rollout)
- abort_argo_rollout (Abort rollout)
- pause_argo_rollout (Pause rollout)
- set_argo_rollout_weight (Set canary weight)
- set_argo_rollout_image (Update rollout image)

### Helm Tools
- helm_list_releases (List Helm releases)
- helm_get_release (Get release details)
- helm_install_release (Install release)
- helm_upgrade_release (Upgrade release)
- helm_uninstall_release (Uninstall release)
- helm_list_repositories (List repositories)
- helm_add_repository (Add repository)
- helm_update_repositories (Update repositories)

### Converters
- docker_compose_to_k8s_manifest (Convert Docker Compose to K8s)
- k8s_manifest_to_helm_chart (Convert K8s to Helm)
- k8s_manifest_to_argo_rollout (Convert Deployment to Argo Rollout)

## Example Prompts

### Example 1: Kubernetes Event Analysis

Input: "Please generate me a prompt for kubernetes events to debug issue"

Output:
"""
Analyze Kubernetes events to debug issues, identify patterns, and provide actionable recommendations.

Parameters:
	Event Type: ${event_type} (If provided, focus on this type of event)
	Namespace: ${namespace} (If provided, focus on events in this namespace)
	Resource Name: ${resource_name} (If provided, focus on events related to this resource)
	Resource Kind: ${resource_kind} (If provided, focus on events related to this resource kind)
	Time Range: ${time_range} (If provided, focus on events within this time range; default is last 1h)

Please perform these steps:
- First, determine the scope of the event analysis based on provided parameters:
	If Event Type is provided, filter events by this type (e.g., Warning, Normal, Error)
	If Namespace is provided, focus on events in this specific namespace
	If Resource Name is provided, focus on events involving this specific resource
	If Resource Kind is provided, focus on events involving this kind of resource (e.g., Pod, Deployment)
	If Time Range is provided, focus on events within that time range; otherwise default to last hour
- Gather relevant events based on the determined scope using the events_list tool with appropriate filters.
- For any resources involved in significant events, gather additional context:
	For pod-related events: check pod status, logs, and configuration
	For node-related events: check node status and metrics
	For deployment/statefulset events: check rollout status and history
	For service/ingress events: check endpoint availability
- Analyze the collected events to:
	Identify patterns and recurring issues
	Determine potential root causes
	Establish a timeline of related events
	Correlate events with resource state changes
- Based on the collected data, please provide:
	Summary of significant events, sorted by severity and recency
	Analysis of event patterns and their implications
	Likely root causes for observed issues
	Recommended actions to resolve identified problems
	Preventive measures to avoid similar issues in the future

Use these tools to perform the analysis:
	events_list (to gather events with appropriate filters)
	pods_get, nodes_get, etc. (to get details about resources mentioned in events)
	pods_log (to check logs for pods involved in events)
	resources_get (to examine configuration of resources involved)
	prometheus_metrics_query (if available, to correlate events with metrics)

Present the results in a clear, organized format with separate sections for:
	Event summary and classification
	Detailed event analysis with context
	Root cause analysis if related logs are provided
	Recommended actions
	Preventive measures
	Evidence sections
	App or infra problem determination

For critical events, highlight them clearly and provide immediate mitigation steps if applicable.
"""

## Instructions

When generating a prompt based on the user's description:

1. Identify the core task being requested
2. Select appropriate parameters that would be needed for the task
3. Create a logical sequence of steps to perform the task
4. Include only the most relevant tools from the above list
5. Structure the presentation of results in a clear, organized way
6. Include specific metrics or data points that would be valuable
7. Format the response in a clean, readable manner with appropriate spacing

Remember to keep the generated prompt comprehensive but focused, and ensure it follows a logical workflow that addresses the specific task requested by the user.
`
