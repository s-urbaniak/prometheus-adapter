package v1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResourceRule describes a rule for querying resource metrics API results.
// The name of the resource rule reflects the actual resource, i.e. cpu, memory.
type ResourceRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ResourceRuleSpec
}

// ResourceRuleList is a list of ResourceRules.
type ResourceRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []*ResourceRule `json:"items"`
}

type ResourceRuleSpec struct {
	NodeQuery string                   `json:"nodeQuery"`
	PodQuery  string                   `json:"podQuery"`
	Labels    map[string]GroupResource `json:"labels"`
	Window    time.Duration            `json:"window"`
}

type GroupResource struct {
	Group    string `json:"group,omitempty"`
	Resource string `json:"resource"`
}
