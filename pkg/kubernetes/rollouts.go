package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ResourceRollout performs rollout operations on Kubernetes resources
func (k *Kubernetes) ResourceRollout(ctx context.Context, namespace, resourceType, resourceName, action string, revision int) (string, error) {
	resourceType = strings.ToLower(resourceType)
	action = strings.ToLower(action)

	// Ensure namespace is set
	namespace = namespaceOrDefault(namespace)

	// Handle different actions
	switch action {
	case "history":
		return k.getRolloutHistory(ctx, namespace, resourceType, resourceName)
	case "pause":
		return k.pauseRollout(ctx, namespace, resourceType, resourceName)
	case "resume":
		return k.resumeRollout(ctx, namespace, resourceType, resourceName)
	case "restart":
		return k.restartRollout(ctx, namespace, resourceType, resourceName)
	case "status":
		return k.getRolloutStatus(ctx, namespace, resourceType, resourceName)
	case "undo":
		return k.undoRollout(ctx, namespace, resourceType, resourceName, revision)
	default:
		return "", fmt.Errorf("unsupported rollout action: %s", action)
	}
}

// getRolloutHistory retrieves the rollout history of a resource
func (k *Kubernetes) getRolloutHistory(ctx context.Context, namespace, resourceType, resourceName string) (string, error) {
	switch resourceType {
	case "deployment", "deploy":
		deploy, err := k.clientSet.AppsV1().Deployments(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		// Get ReplicaSets owned by this Deployment
		selector, err := metav1.LabelSelectorAsSelector(deploy.Spec.Selector)
		if err != nil {
			return "", err
		}

		rsList, err := k.clientSet.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: selector.String(),
		})
		if err != nil {
			return "", err
		}

		// Build history information
		historyInfo := fmt.Sprintf("REVISION  CHANGE-CAUSE\n")
		for _, rs := range rsList.Items {
			// Check if this ReplicaSet is owned by the Deployment
			isOwned := false
			for _, ownerRef := range rs.OwnerReferences {
				if ownerRef.Kind == "Deployment" && ownerRef.Name == resourceName {
					isOwned = true
					break
				}
			}

			if isOwned {
				revision := rs.Annotations["deployment.kubernetes.io/revision"]
				changeCause := rs.Annotations["kubernetes.io/change-cause"]
				if changeCause == "" {
					changeCause = "<none>"
				}
				historyInfo += fmt.Sprintf("%-9s %s\n", revision, changeCause)
			}
		}
		return historyInfo, nil

	case "statefulset", "sts":
		// Similar implementation for StatefulSets
		sts, err := k.clientSet.AppsV1().StatefulSets(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("StatefulSet %s/%s has revision %s\n", namespace, resourceName, sts.Status.UpdateRevision), nil

	case "daemonset", "ds":
		// Similar implementation for DaemonSets
		_, err := k.clientSet.AppsV1().DaemonSets(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("DaemonSet %s/%s has no rollout history\n", namespace, resourceName), nil

	default:
		return "", fmt.Errorf("unsupported resource type for rollout history: %s", resourceType)
	}
}

// pauseRollout pauses a rollout
func (k *Kubernetes) pauseRollout(ctx context.Context, namespace, resourceType, resourceName string) (string, error) {
	switch resourceType {
	case "deployment", "deploy":
		// Patch to set .spec.paused to true
		patchBytes := []byte(`{"spec":{"paused":true}}`)
		_, err := k.clientSet.AppsV1().Deployments(namespace).Patch(ctx, resourceName, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("deployment.apps/%s paused\n", resourceName), nil

	case "statefulset", "sts":
		// Note: StatefulSets don't have a pause feature in the same way Deployments do
		return "", fmt.Errorf("rollout pause is not supported for statefulsets")

	case "daemonset", "ds":
		// Note: DaemonSets don't have a pause feature in the same way Deployments do
		return "", fmt.Errorf("rollout pause is not supported for daemonsets")

	default:
		return "", fmt.Errorf("unsupported resource type for rollout pause: %s", resourceType)
	}
}

// resumeRollout resumes a paused rollout
func (k *Kubernetes) resumeRollout(ctx context.Context, namespace, resourceType, resourceName string) (string, error) {
	switch resourceType {
	case "deployment", "deploy":
		// Patch to set .spec.paused to false
		patchBytes := []byte(`{"spec":{"paused":false}}`)
		_, err := k.clientSet.AppsV1().Deployments(namespace).Patch(ctx, resourceName, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("deployment.apps/%s resumed\n", resourceName), nil

	case "statefulset", "sts":
		// Note: StatefulSets don't have a pause feature in the same way Deployments do
		return "", fmt.Errorf("rollout resume is not supported for statefulsets")

	case "daemonset", "ds":
		// Note: DaemonSets don't have a pause feature in the same way Deployments do
		return "", fmt.Errorf("rollout resume is not supported for daemonsets")

	default:
		return "", fmt.Errorf("unsupported resource type for rollout resume: %s", resourceType)
	}
}

// restartRollout restarts a rollout
func (k *Kubernetes) restartRollout(ctx context.Context, namespace, resourceType, resourceName string) (string, error) {
	switch resourceType {
	case "deployment", "deploy":
		// Common way to restart a deployment is to add a restart annotation
		// First get the current deployment
		deploy, err := k.clientSet.AppsV1().Deployments(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		// Create a copy of the deployment annotations or initialize if nil
		annotations := map[string]string{}
		if deploy.Spec.Template.Annotations != nil {
			for k, v := range deploy.Spec.Template.Annotations {
				annotations[k] = v
			}
		}

		// Add or update the restart timestamp annotation
		annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

		// Create patch with updated annotations
		patchBytes := []byte(fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":%s}}}}`, marshalAnnotations(annotations)))
		_, err = k.clientSet.AppsV1().Deployments(namespace).Patch(ctx, resourceName, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("deployment.apps/%s restarted\n", resourceName), nil

	case "statefulset", "sts":
		// Similar approach for StatefulSets
		sts, err := k.clientSet.AppsV1().StatefulSets(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		annotations := map[string]string{}
		if sts.Spec.Template.Annotations != nil {
			for k, v := range sts.Spec.Template.Annotations {
				annotations[k] = v
			}
		}
		annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

		patchBytes := []byte(fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":%s}}}}`, marshalAnnotations(annotations)))
		_, err = k.clientSet.AppsV1().StatefulSets(namespace).Patch(ctx, resourceName, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("statefulset.apps/%s restarted\n", resourceName), nil

	case "daemonset", "ds":
		// Similar approach for DaemonSets
		ds, err := k.clientSet.AppsV1().DaemonSets(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		annotations := map[string]string{}
		if ds.Spec.Template.Annotations != nil {
			for k, v := range ds.Spec.Template.Annotations {
				annotations[k] = v
			}
		}
		annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

		patchBytes := []byte(fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":%s}}}}`, marshalAnnotations(annotations)))
		_, err = k.clientSet.AppsV1().DaemonSets(namespace).Patch(ctx, resourceName, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("daemonset.apps/%s restarted\n", resourceName), nil

	default:
		return "", fmt.Errorf("unsupported resource type for rollout restart: %s", resourceType)
	}
}

// getRolloutStatus checks the status of a rollout
func (k *Kubernetes) getRolloutStatus(ctx context.Context, namespace, resourceType, resourceName string) (string, error) {
	switch resourceType {
	case "deployment", "deploy":
		deploy, err := k.clientSet.AppsV1().Deployments(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		// Check Deployment status
		status := fmt.Sprintf("deployment %s/%s\n", namespace, resourceName)
		status += fmt.Sprintf("Desired: %d, Updated: %d, Total: %d, Available: %d, Ready: %d\n",
			deploy.Status.Replicas,
			deploy.Status.UpdatedReplicas,
			deploy.Status.Replicas,
			deploy.Status.AvailableReplicas,
			deploy.Status.ReadyReplicas)

		if deploy.Spec.Paused {
			status += "Deployment is paused\n"
		}

		if deploy.Status.UpdatedReplicas < deploy.Status.Replicas {
			status += fmt.Sprintf("Waiting for rollout to finish: %d out of %d new replicas have been updated...\n",
				deploy.Status.UpdatedReplicas, deploy.Status.Replicas)
		} else if deploy.Status.AvailableReplicas < deploy.Status.UpdatedReplicas {
			status += fmt.Sprintf("Waiting for rollout to finish: %d of %d updated replicas are available...\n",
				deploy.Status.AvailableReplicas, deploy.Status.UpdatedReplicas)
		} else {
			status += "Deployment successfully rolled out\n"
		}
		return status, nil

	case "statefulset", "sts":
		sts, err := k.clientSet.AppsV1().StatefulSets(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		status := fmt.Sprintf("statefulset %s/%s\n", namespace, resourceName)
		status += fmt.Sprintf("Desired: %d, Current: %d, Ready: %d, Updated: %d\n",
			sts.Status.Replicas,
			sts.Status.CurrentReplicas,
			sts.Status.ReadyReplicas,
			sts.Status.UpdatedReplicas)

		if sts.Status.UpdatedReplicas < sts.Status.Replicas {
			status += fmt.Sprintf("Waiting for %d pods to be updated...\n",
				sts.Status.Replicas-sts.Status.UpdatedReplicas)
		} else if sts.Status.ReadyReplicas < sts.Status.Replicas {
			status += fmt.Sprintf("Waiting for %d pods to be ready...\n",
				sts.Status.Replicas-sts.Status.ReadyReplicas)
		} else {
			status += "StatefulSet rolling update complete\n"
		}
		return status, nil

	case "daemonset", "ds":
		ds, err := k.clientSet.AppsV1().DaemonSets(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		status := fmt.Sprintf("daemonset %s/%s\n", namespace, resourceName)
		status += fmt.Sprintf("Desired: %d, Current: %d, Ready: %d, Updated: %d, Available: %d\n",
			ds.Status.DesiredNumberScheduled,
			ds.Status.CurrentNumberScheduled,
			ds.Status.NumberReady,
			ds.Status.UpdatedNumberScheduled,
			ds.Status.NumberAvailable)

		if ds.Status.UpdatedNumberScheduled < ds.Status.DesiredNumberScheduled {
			status += fmt.Sprintf("Waiting for %d pods to be updated...\n",
				ds.Status.DesiredNumberScheduled-ds.Status.UpdatedNumberScheduled)
		} else if ds.Status.NumberReady < ds.Status.DesiredNumberScheduled {
			status += fmt.Sprintf("Waiting for %d pods to be ready...\n",
				ds.Status.DesiredNumberScheduled-ds.Status.NumberReady)
		} else {
			status += "DaemonSet rolling update complete\n"
		}
		return status, nil

	default:
		return "", fmt.Errorf("unsupported resource type for rollout status: %s", resourceType)
	}
}

// undoRollout rolls back a resource to a previous version
func (k *Kubernetes) undoRollout(ctx context.Context, namespace, resourceType, resourceName string, revision int) (string, error) {
	switch resourceType {
	case "deployment", "deploy":
		// Get the current deployment
		deploy, err := k.clientSet.AppsV1().Deployments(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		// Get ReplicaSets owned by this Deployment to find the revision
		selector, err := metav1.LabelSelectorAsSelector(deploy.Spec.Selector)
		if err != nil {
			return "", err
		}

		rsList, err := k.clientSet.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: selector.String(),
		})
		if err != nil {
			return "", err
		}

		// Find target revision ReplicaSet
		var targetRS *appsv1.ReplicaSet
		var targetRevision int = revision

		if revision > 0 {
			// Find specific revision
			for i := range rsList.Items {
				rs := &rsList.Items[i]
				// Check if this RS is owned by the deployment
				isOwned := false
				for _, ownerRef := range rs.OwnerReferences {
					if ownerRef.Kind == "Deployment" && ownerRef.Name == resourceName {
						isOwned = true
						break
					}
				}

				if isOwned {
					revisionStr := rs.Annotations["deployment.kubernetes.io/revision"]
					rsRevision, err := strconv.Atoi(revisionStr)
					if err != nil {
						continue
					}
					if rsRevision == revision {
						targetRS = rs
						break
					}
				}
			}
			if targetRS == nil {
				return "", fmt.Errorf("couldn't find revision %d of deployment %s/%s", revision, namespace, resourceName)
			}
		} else {
			// Find previous revision if none specified
			// Sort ReplicaSets by revision (higher is newer)
			// Skip the current one and get the previous one
			currentRevision := deploy.Annotations["deployment.kubernetes.io/revision"]
			currentRevInt, err := strconv.Atoi(currentRevision)
			if err != nil {
				currentRevInt = 0
			}

			prevRevInt := 0
			for i := range rsList.Items {
				rs := &rsList.Items[i]
				// Check if this RS is owned by the deployment
				isOwned := false
				for _, ownerRef := range rs.OwnerReferences {
					if ownerRef.Kind == "Deployment" && ownerRef.Name == resourceName {
						isOwned = true
						break
					}
				}

				if isOwned {
					revisionStr := rs.Annotations["deployment.kubernetes.io/revision"]
					rsRevision, err := strconv.Atoi(revisionStr)
					if err != nil {
						continue
					}
					if rsRevision < currentRevInt && rsRevision > prevRevInt {
						prevRevInt = rsRevision
						targetRS = rs
						targetRevision = prevRevInt
					}
				}
			}
			if targetRS == nil {
				return "", fmt.Errorf("no revision prior to current revision found for deployment %s/%s", namespace, resourceName)
			}
		}

		// Create a new deployment based on the current one, but with the old template from the target ReplicaSet
		deploymentCopy := deploy.DeepCopy()

		// Update the deployment with the template from the target ReplicaSet
		deploymentCopy.Spec.Template = targetRS.Spec.Template

		// Remove some fields that shouldn't be copied directly
		if deploymentCopy.Spec.Template.ObjectMeta.Labels != nil {
			delete(deploymentCopy.Spec.Template.ObjectMeta.Labels, "pod-template-hash")
		}

		// Add rollback annotation
		if deploymentCopy.Annotations == nil {
			deploymentCopy.Annotations = make(map[string]string)
		}
		deploymentCopy.Annotations["kubernetes.io/change-cause"] = fmt.Sprintf("kubectl rollout undo deployment/%s to revision %d", resourceName, targetRevision)

		// Update the deployment
		_, err = k.clientSet.AppsV1().Deployments(namespace).Update(ctx, deploymentCopy, metav1.UpdateOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to rollback deployment: %v", err)
		}

		return fmt.Sprintf("deployment.apps/%s rolled back to revision %d\n", resourceName, targetRevision), nil

	case "statefulset", "sts":
		// Note: StatefulSets have limited rollback capabilities
		return "", fmt.Errorf("rollout undo is not fully supported for statefulsets yet")

	case "daemonset", "ds":
		// Note: DaemonSets have limited rollback capabilities
		return "", fmt.Errorf("rollout undo is not fully supported for daemonsets yet")

	default:
		return "", fmt.Errorf("unsupported resource type for rollout undo: %s", resourceType)
	}
}

// marshalAnnotations is a helper to format annotations for a patch
func marshalAnnotations(annotations map[string]string) string {
	if len(annotations) == 0 {
		return "{}"
	}

	parts := make([]string, 0, len(annotations))
	for k, v := range annotations {
		parts = append(parts, fmt.Sprintf(`"%s":"%s"`, k, v))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

// marshalPodSpec is a helper to format pod spec for a patch
func marshalPodSpec(podSpec interface{}) []byte {
	// Properly marshal the pod spec
	data, err := json.Marshal(podSpec)
	if err != nil {
		// If marshaling fails, return an empty object rather than failing
		return []byte("{}")
	}
	return data
}
