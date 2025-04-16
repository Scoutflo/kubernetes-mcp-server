package kubernetes

import (
	"encoding/json"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type PrometheusRuleSpec struct {
	Groups []RuleGroup `json:"groups"`
}

type RuleGroup struct {
	Name                    string `json:"name"`
	Interval                string `json:"interval,omitempty"`
	Limit                   *int   `json:"limit,omitempty"`
	PartialResponseStrategy string `json:"partial_response_strategy,omitempty"`
	QueryOffset             string `json:"query_offset,omitempty"`
	Rules                   []Rule `json:"rules"`
}

type Rule struct {
	Alert         string            `json:"alert,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	Expr          string            `json:"expr"` // Changed from runtime.RawExtension to string
	For           string            `json:"for,omitempty"`
	KeepFiringFor string            `json:"keep_firing_for,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Record        string            `json:"record,omitempty"`
}

type PrometheusRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PrometheusRuleSpec `json:"spec"`
}

type PrometheusRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PrometheusRule `json:"items"`
}

type PrometheusCreateAlertApiRequest struct {
	AlertName   string            `json:"alert_name" binding:"required,min=2,max=100"`
	Interval    string            `json:"interval,omitempty"`
	Expression  string            `json:"expression" binding:"required"`
	AppLabel    string            `json:"app_label" binding:"required"`
	For         string            `json:"for"`
	Namespace   string            `json:"namespace" binding:"required"`
	Annotations map[string]string `json:"annotations,omitempty"`
	AlertLabels map[string]string `json:"alert_labels,omitempty"`
}

type PrometheusUpdateAlertApiRequest struct {
	AlertName   string            `json:"alert_name,omitempty"`
	Interval    string            `json:"interval,omitempty"`
	Expression  string            `json:"expression,omitempty"`
	AppLabel    string            `json:"app_label,omitempty"`
	For         string            `json:"for,omitempty"`
	Namespace   string            `json:"namespace" binding:"required"`
	Annotations map[string]string `json:"annotations,omitempty"`
	AlertLabels map[string]string `json:"alert_labels,omitempty"`
}

type PrometheusDeleteAlertApiRequest struct {
	AppLabel  string `json:"app_label" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	AlertName string `json:"alert_name"`
}

// CreatePrometheusAlert creates a new Prometheus alert rule
func (k *Kubernetes) CreatePrometheusAlert(alertName, expression, appLabel, namespace, interval, forDuration string, annotations, alertLabels map[string]string) (string, error) {
	// Initialize labels if nil
	if alertLabels == nil {
		alertLabels = make(map[string]string)
	}

	// Set standard labels
	alertLabels["app"] = appLabel
	alertLabels["name"] = alertName

	// Check if the rule already exists
	existingPrometheusRule, err := k.GetPrometheusRuleInstance(appLabel, namespace)
	if err != nil {
		// If error is not "not found", return it
		if !isNotFoundError(err) {
			return "", fmt.Errorf("failed to get PrometheusRule: %v", err)
		}
		// Create new prometheus rule instance if not found
		prometheusRule := &PrometheusRule{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "monitoring.coreos.com/v1",
				Kind:       "PrometheusRule",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      appLabel,
				Namespace: namespace,
				Labels:    map[string]string{"release": "kube-prometheus-stack", "app": appLabel},
			},
			Spec: PrometheusRuleSpec{
				Groups: []RuleGroup{
					{
						Name:     appLabel,
						Interval: interval,
						Rules: []Rule{
							{
								Alert:       alertName,
								Expr:        expression,
								For:         forDuration,
								Labels:      alertLabels,
								Annotations: annotations,
							},
						},
					},
				},
			},
		}

		// Create the PrometheusRule instance
		err = k.CreatePrometheusRuleInstance(prometheusRule)
		if err != nil {
			return "", fmt.Errorf("failed to create PrometheusRule: %v", err)
		}
	} else {
		// Alert rule exists, check if this alert name already exists
		alertExists := false
		for _, group := range existingPrometheusRule.Spec.Groups {
			for _, rule := range group.Rules {
				if rule.Alert == alertName {
					alertExists = true
					break
				}
			}
			if alertExists {
				break
			}
		}

		if alertExists {
			return "", fmt.Errorf("alert '%s' already exists in rule group '%s'", alertName, appLabel)
		}

		// Append the alert to the existing prometheus rule
		request := PrometheusCreateAlertApiRequest{
			AlertName:   alertName,
			Expression:  expression,
			AppLabel:    appLabel,
			For:         forDuration,
			Namespace:   namespace,
			Annotations: annotations,
			AlertLabels: alertLabels,
			Interval:    interval,
		}

		updatedRule := k.AppendAlertInPrometheusRuleInstanceStruct(request, existingPrometheusRule)

		// Update the PrometheusRule instance
		err = k.UpdatePrometheusRuleInstance(updatedRule)
		if err != nil {
			return "", fmt.Errorf("failed to update PrometheusRule: %v", err)
		}
	}

	// Get the updated rule for response
	updatedPrometheusRule, err := k.GetPrometheusRuleInstance(appLabel, namespace)
	if err != nil {
		return "", fmt.Errorf("failed to get updated PrometheusRule: %v", err)
	}

	// Convert to JSON
	ruleJSON, err := json.MarshalIndent(updatedPrometheusRule, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal PrometheusRule: %v", err)
	}

	return fmt.Sprintf("Alert '%s' created successfully in namespace '%s'.\n\nAlert definition:\n%s",
		alertName, namespace, string(ruleJSON)), nil
}

// UpdatePrometheusAlert updates an existing Prometheus alert rule
func (k *Kubernetes) UpdatePrometheusAlert(alertName, expression, appLabel, namespace, interval, forDuration string, annotations, alertLabels map[string]string) (string, error) {
	// Check if the PrometheusRule exists
	existingPrometheusRule, err := k.GetPrometheusRuleInstance(appLabel, namespace)
	if err != nil {
		if isNotFoundError(err) {
			// If not found and we have all required parameters, create a new rule
			if alertName != "" && expression != "" {
				return k.CreatePrometheusAlert(alertName, expression, appLabel, namespace, interval, forDuration, annotations, alertLabels)
			}
			return "", fmt.Errorf("PrometheusRule '%s' not found in namespace '%s'", appLabel, namespace)
		}
		return "", fmt.Errorf("failed to get PrometheusRule: %v", err)
	}

	// Check if the alert exists in the rule
	alertExists := false
	for _, group := range existingPrometheusRule.Spec.Groups {
		for _, rule := range group.Rules {
			if rule.Alert == alertName {
				alertExists = true
				break
			}
		}
		if alertExists {
			break
		}
	}

	if !alertExists {
		return "", fmt.Errorf("alert '%s' not found in rule group '%s'", alertName, appLabel)
	}

	// Update the alert in the existing PrometheusRule
	updateRequest := PrometheusUpdateAlertApiRequest{
		AlertName:   alertName,
		Expression:  expression,
		AppLabel:    appLabel,
		For:         forDuration,
		Namespace:   namespace,
		Annotations: annotations,
		AlertLabels: alertLabels,
		Interval:    interval,
	}

	updatedRule := UpdateAlertInPrometheusRuleInstanceStruct(updateRequest, existingPrometheusRule)

	// Update the PrometheusRule instance
	err = k.UpdatePrometheusRuleInstance(updatedRule)
	if err != nil {
		return "", fmt.Errorf("failed to update PrometheusRule: %v", err)
	}

	// Get the updated rule for response
	updatedPrometheusRule, err := k.GetPrometheusRuleInstance(appLabel, namespace)
	if err != nil {
		return "", fmt.Errorf("failed to get updated PrometheusRule: %v", err)
	}

	// Convert to JSON
	updateJSON, err := json.MarshalIndent(updatedPrometheusRule, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal PrometheusRule: %v", err)
	}

	return fmt.Sprintf("Alert '%s' updated successfully in namespace '%s'.\n\nUpdate details:\n%s",
		alertName, namespace, string(updateJSON)), nil
}

// DeletePrometheusAlert deletes a Prometheus alert rule
func (k *Kubernetes) DeletePrometheusAlert(appLabel, namespace, alertName string) (string, error) {
	// Check if the PrometheusRule exists
	existingPrometheusRule, err := k.GetPrometheusRuleInstance(appLabel, namespace)
	if err != nil {
		if isNotFoundError(err) {
			return "", fmt.Errorf("PrometheusRule '%s' not found in namespace '%s'", appLabel, namespace)
		}
		return "", fmt.Errorf("failed to get PrometheusRule: %v", err)
	}

	// If alertName is specified, remove just that alert
	if alertName != "" {
		// Check if the alert exists in the rule
		alertExists := false
		for _, group := range existingPrometheusRule.Spec.Groups {
			for _, rule := range group.Rules {
				if rule.Alert == alertName {
					alertExists = true
					break
				}
			}
			if alertExists {
				break
			}
		}

		if !alertExists {
			return "", fmt.Errorf("alert '%s' not found in rule group '%s'", alertName, appLabel)
		}

		// Remove the specific alert
		deleteRequest := PrometheusDeleteAlertApiRequest{
			AppLabel:  appLabel,
			Namespace: namespace,
			AlertName: alertName,
		}

		updatedRule := RemoveAlertInPrometheusRuleInstanceStruct(deleteRequest, existingPrometheusRule)

		// Check if there are any alerts left
		anyAlertsLeft := false
		for _, group := range updatedRule.Spec.Groups {
			if len(group.Rules) > 0 {
				anyAlertsLeft = true
				break
			}
		}

		// If no alerts are left, delete the entire rule
		if !anyAlertsLeft {
			err = k.DeletePrometheusRuleAlertInstance(appLabel, namespace)
			if err != nil {
				return "", fmt.Errorf("failed to delete empty PrometheusRule: %v", err)
			}
		} else {
			// Update the PrometheusRule instance
			err = k.UpdatePrometheusRuleInstance(updatedRule)
			if err != nil {
				return "", fmt.Errorf("failed to update PrometheusRule: %v", err)
			}
		}

		result := map[string]interface{}{
			"operation": "delete",
			"alertName": alertName,
			"appLabel":  appLabel,
			"namespace": namespace,
			"scope":     "specific alert",
		}

		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal result info: %v", err)
		}

		return fmt.Sprintf("Alert '%s' deleted successfully from rule group '%s' in namespace '%s'.\n\nDeletion details:\n%s",
			alertName, appLabel, namespace, string(resultJSON)), nil
	} else {
		// Delete the entire PrometheusRule
		err = k.DeletePrometheusRuleAlertInstance(appLabel, namespace)
		if err != nil {
			return "", fmt.Errorf("failed to delete PrometheusRule: %v", err)
		}

		result := map[string]interface{}{
			"operation": "delete",
			"appLabel":  appLabel,
			"namespace": namespace,
			"scope":     "entire rule group",
		}

		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal result info: %v", err)
		}

		return fmt.Sprintf("Entire rule group '%s' deleted successfully from namespace '%s'.\n\nDeletion details:\n%s",
			appLabel, namespace, string(resultJSON)), nil
	}
}

func (k *Kubernetes) GetPrometheusRuleInstance(appLabel string, namespace string) (*PrometheusRule, error) {
	resourceObject, err := k.GetCrdResource("monitoringPrometheusRule", appLabel, namespace)
	if err != nil {
		return nil, err
	}
	var prometheusRule *PrometheusRule
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(resourceObject.Object, &prometheusRule); err != nil {
		return nil, err
	}
	return prometheusRule, nil
}

func UpdateAlertInPrometheusRuleInstanceStruct(prometheusApiRequest PrometheusUpdateAlertApiRequest, prometheusRule *PrometheusRule) *PrometheusRule {
	for groupIndex, group := range prometheusRule.Spec.Groups {
		for ruleIndex, rule := range group.Rules {
			if rule.Alert == prometheusApiRequest.AlertName {
				if prometheusApiRequest.Interval != "" {
					prometheusRule.Spec.Groups[groupIndex].Interval = prometheusApiRequest.Interval
				}
				if prometheusApiRequest.Expression != "" {
					prometheusRule.Spec.Groups[groupIndex].Rules[ruleIndex].Expr = prometheusApiRequest.Expression
				}
				if prometheusApiRequest.Annotations != nil {
					prometheusRule.Spec.Groups[groupIndex].Rules[ruleIndex].Annotations = prometheusApiRequest.Annotations
				}
				if prometheusApiRequest.AlertLabels != nil {
					existingLabels := prometheusRule.Spec.Groups[groupIndex].Rules[ruleIndex].Labels
					for key, value := range prometheusApiRequest.AlertLabels {
						existingLabels[key] = value
					}
					prometheusRule.Spec.Groups[groupIndex].Rules[ruleIndex].Labels = existingLabels
				}
				if prometheusApiRequest.For != "" {
					prometheusRule.Spec.Groups[groupIndex].Rules[ruleIndex].For = prometheusApiRequest.For
				}
			}
		}
	}
	return prometheusRule
}

func RemoveAlertInPrometheusRuleInstanceStruct(prometheusApiRequest PrometheusDeleteAlertApiRequest, prometheusRule *PrometheusRule) *PrometheusRule {
	for groupIndex, group := range prometheusRule.Spec.Groups {
		var newRules []Rule
		for _, rule := range group.Rules {
			if rule.Alert != prometheusApiRequest.AlertName {
				newRules = append(newRules, rule)
			}
		}
		prometheusRule.Spec.Groups[groupIndex].Rules = newRules
	}

	return prometheusRule
}

func (k *Kubernetes) CreatePrometheusRuleInstance(prometheusRule *PrometheusRule) error {
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(prometheusRule)
	if err != nil {
		return err
	}
	err = k.CreateCrdResource("monitoringPrometheusRule", unstructuredObj, prometheusRule.ObjectMeta.Namespace)
	if err != nil {
		return err
	}
	return nil
}

func (k *Kubernetes) AppendAlertInPrometheusRuleInstanceStruct(prometheusApiRequest PrometheusCreateAlertApiRequest, prometheusRule *PrometheusRule) *PrometheusRule {
	prometheusApiRequest.AlertLabels["app"] = prometheusApiRequest.AppLabel
	prometheusApiRequest.AlertLabels["name"] = prometheusApiRequest.AlertName
	newRule := Rule{
		Alert:       prometheusApiRequest.AlertName,
		Expr:        prometheusApiRequest.Expression,
		For:         prometheusApiRequest.For,
		Labels:      prometheusApiRequest.AlertLabels,
		Annotations: prometheusApiRequest.Annotations,
	}
	prometheusRule.Spec.Groups[0].Rules = append(prometheusRule.Spec.Groups[0].Rules, newRule)
	return prometheusRule
}

func (k *Kubernetes) UpdatePrometheusRuleInstance(prometheusRule *PrometheusRule) error {
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(prometheusRule)
	if err != nil {
		return err
	}
	err = k.UpdateCrdResource("monitoringPrometheusRule", unstructuredObj, prometheusRule.ObjectMeta.Namespace)
	if err != nil {
		return err
	}
	return nil
}

func (k *Kubernetes) DeletePrometheusRuleAlertInstance(appLabel string, namespace string) error {
	err := k.DeleteCrdResource("monitoringPrometheusRule", appLabel, namespace)
	return err
}

// isNotFoundError checks if an error is a "not found" Kubernetes error
func isNotFoundError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}
