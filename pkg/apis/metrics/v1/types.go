/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceRules = "resourcerules"
)

// ResourceRule describes a rule for querying resource metrics API results.
// The name of the resource rule reflects the actual resource, i.e. cpu, memory.
//
// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type ResourceRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ResourceRuleSpec
}

// ResourceRuleList is a list of ResourceRules.
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type ResourceRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []*ResourceRule `json:"items"`
}

// +k8s:openapi-gen=true
type ResourceRuleSpec struct {
	NodeQuery string                   `json:"nodeQuery"`
	PodQuery  string                   `json:"podQuery"`
	Labels    map[string]GroupResource `json:"labels"`
	Window    metav1.Duration          `json:"window"`
}

// +k8s:openapi-gen=true
type GroupResource struct {
	Group    string `json:"group,omitempty"`
	Resource string `json:"resource"`
}
