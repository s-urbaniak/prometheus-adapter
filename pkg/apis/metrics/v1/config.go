package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ResourceMetrics struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ResourceMetricsSpec `json:"spec"`
}

// +k8s:openapi-gen=true
type ResourceMetricsSpec struct {
	Rules  ResourceMetricsRules `json:"rules"`
	Window metav1.Duration      `json:"window"`
}

// +k8s:openapi-gen=true
type ResourceMetricsRules struct {
	CPU    ResourceMetricsRule `json:"cpu"`
	Memory ResourceMetricsRule `json:"memory"`
}

// +k8s:openapi-gen=true
type ResourceMetricsRule struct {
	Queries        Queries                  `json:"queries"`
	Labels         map[string]GroupResource `json:"labels"`
	ContainerLabel string                   `json:"containerLabel"`
}

// +k8s:openapi-gen=true
type Queries struct {
	Node string `json:"node"`
	Pod  string `json:"pod"`
}
