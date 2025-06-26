package kubernetes

import (
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
