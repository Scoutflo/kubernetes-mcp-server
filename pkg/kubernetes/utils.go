package kubernetes

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var resourceMap = map[string]schema.GroupVersionResource{
	"namespace":                      {Group: "", Version: "v1", Resource: "namespaces"},
	"node":                           {Group: "", Version: "v1", Resource: "nodes"},
	"configmap":                      {Group: "", Version: "v1", Resource: "configmaps"},
	"persistentvolume":               {Group: "", Version: "v1", Resource: "persistentvolumes"},
	"persistentvolumeclaim":          {Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
	"poddisruptionbudget":            {Group: "policy", Version: "v1", Resource: "poddisruptionbudgets"},
	"secret":                         {Group: "", Version: "v1", Resource: "secrets"},
	"storageclass":                   {Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"},
	"endpoint":                       {Group: "", Version: "v1", Resource: "endpoints"},
	"ingress":                        {Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
	"ingressclass":                   {Group: "networking.k8s.io", Version: "v1", Resource: "ingressclasses"},
	"networkpolicy":                  {Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"},
	"portforwarding":                 {Group: "", Version: "v1", Resource: "pods/portforward"},
	"service":                        {Group: "", Version: "v1", Resource: "services"},
	"cluster":                        {Group: "", Version: "v1", Resource: "nodes"},
	"clusterIssue":                   {Group: "clusterissue.k8s.io", Version: "v1", Resource: "clusterissues"},
	"crd":                            {Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"},
	"event":                          {Group: "", Version: "v1", Resource: "events"},
	"hpa":                            {Group: "autoscaling", Version: "v2", Resource: "horizontalpodautoscalers"},
	"lease":                          {Group: "coordination.k8s.io", Version: "v1", Resource: "leases"},
	"limitrange":                     {Group: "", Version: "v1", Resource: "limitranges"},
	"mutatingwebhookconfiguration":   {Group: "admissionregistration.k8s.io", Version: "v1", Resource: "mutatingwebhookconfigurations"},
	"priorityclass":                  {Group: "scheduling.k8s.io", Version: "v1", Resource: "priorityclasses"},
	"resourcequota":                  {Group: "", Version: "v1", Resource: "resourcequotas"},
	"validatingwebhookconfiguration": {Group: "admissionregistration.k8s.io", Version: "v1", Resource: "validatingwebhookconfigurations"},
	"clusterrole":                    {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
	"clusterrolebinding":             {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"},
	"role":                           {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},
	"rolebinding":                    {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},
	"serviceaccount":                 {Group: "", Version: "v1", Resource: "serviceaccounts"},
	"storage":                        {Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"},
	"cronjob":                        {Group: "batch", Version: "v1", Resource: "cronjobs"},
	"daemonset":                      {Group: "apps", Version: "v1", Resource: "daemonsets"},
	"deployment":                     {Group: "apps", Version: "v1", Resource: "deployments"},
	"job":                            {Group: "batch", Version: "v1", Resource: "jobs"},
	"pod":                            {Group: "", Version: "v1", Resource: "pods"},
	"replicaset":                     {Group: "apps", Version: "v1", Resource: "replicasets"},
	"replicationcontroller":          {Group: "", Version: "v1", Resource: "replicationcontrollers"},
	"statefulset":                    {Group: "apps", Version: "v1", Resource: "statefulsets"},
	"metrics":                        {Group: "metrics.k8s.io", Version: "v1beta1", Resource: "pods"},
	"monitoringPrometheusRule":       {Group: "monitoring.coreos.com", Version: "v1", Resource: "prometheusrules"},
	"monitoringServiceMonitor":       {Group: "monitoring.coreos.com", Version: "v1", Resource: "servicemonitors"},
	"monitoringalertmanagerconfig":   {Group: "monitoring.coreos.com", Version: "v1alpha1", Resource: "alertmanagerconfigs"},
}

// GetGroupVersionResource is a function that returns a GroupVersionResource for a given resource type
func GetGroupVersionResource(resource string) (schema.GroupVersionResource, error) {
	gvr, exists := resourceMap[resource]
	if !exists {
		return schema.GroupVersionResource{}, fmt.Errorf("unsupported resource type: %s", resource)
	}
	return gvr, nil
}

func (k *Kubernetes) CreateCrdResource(resource string, unstructuredObj map[string]interface{}, namespace string) error {
	gvr, exists := resourceMap[resource]
	if !exists {
		return fmt.Errorf("unsupported resource type: %s", resource)
	}

	unstructuredCRD := &unstructured.Unstructured{Object: unstructuredObj}

	_, err := k.dynamicClient.Resource(gvr).Namespace(namespace).Create(context.TODO(), unstructuredCRD, metav1.CreateOptions{})

	return err
}

func (k *Kubernetes) DeleteCrdResource(resource, name, namespace string) error {
	gvr, exists := resourceMap[resource]
	if !exists {
		return fmt.Errorf("unsupported resource type: %s", resource)
	}

	// Call the Delete method on the dynamic client
	err := k.dynamicClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete CRD resource %s: %w", name, err)
	}

	fmt.Printf("CRD resource %s deleted successfully from namespace %s\n", name, namespace)
	return nil
}

func (k *Kubernetes) UpdateCrdResource(resource string, unstructuredObj map[string]interface{}, namespace string) error {
	gvr, exists := resourceMap[resource]
	if !exists {
		return fmt.Errorf("unsupported resource type: %s", resource)
	}

	unstructuredCRD := &unstructured.Unstructured{Object: unstructuredObj}

	_, err := k.dynamicClient.Resource(gvr).Namespace(namespace).Update(context.TODO(), unstructuredCRD, metav1.UpdateOptions{})

	return err
}

func (k *Kubernetes) GetCrdResource(resource string, name string, namespace string) (*unstructured.Unstructured, error) {
	gvr, exists := resourceMap[resource]
	if !exists {
		return nil, fmt.Errorf("unsupported resource type: %s", resource)
	}
	resourceObject, err := k.dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return resourceObject, nil
}
