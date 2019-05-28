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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	metricsprometheusiov1 "github.com/s-urbaniak/prometheus-adapter/pkg/apis/metrics.prometheus.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeResourceRules implements ResourceRuleInterface
type FakeResourceRules struct {
	Fake *FakeMetricsV1
}

var resourcerulesResource = schema.GroupVersionResource{Group: "metrics.prometheus.io", Version: "v1", Resource: "resourcerules"}

var resourcerulesKind = schema.GroupVersionKind{Group: "metrics.prometheus.io", Version: "v1", Kind: "ResourceRule"}

// Get takes name of the resourceRule, and returns the corresponding resourceRule object, and an error if there is any.
func (c *FakeResourceRules) Get(name string, options v1.GetOptions) (result *metricsprometheusiov1.ResourceRule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(resourcerulesResource, name), &metricsprometheusiov1.ResourceRule{})
	if obj == nil {
		return nil, err
	}
	return obj.(*metricsprometheusiov1.ResourceRule), err
}

// List takes label and field selectors, and returns the list of ResourceRules that match those selectors.
func (c *FakeResourceRules) List(opts v1.ListOptions) (result *metricsprometheusiov1.ResourceRuleList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(resourcerulesResource, resourcerulesKind, opts), &metricsprometheusiov1.ResourceRuleList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &metricsprometheusiov1.ResourceRuleList{ListMeta: obj.(*metricsprometheusiov1.ResourceRuleList).ListMeta}
	for _, item := range obj.(*metricsprometheusiov1.ResourceRuleList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested resourceRules.
func (c *FakeResourceRules) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(resourcerulesResource, opts))
}

// Create takes the representation of a resourceRule and creates it.  Returns the server's representation of the resourceRule, and an error, if there is any.
func (c *FakeResourceRules) Create(resourceRule *metricsprometheusiov1.ResourceRule) (result *metricsprometheusiov1.ResourceRule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(resourcerulesResource, resourceRule), &metricsprometheusiov1.ResourceRule{})
	if obj == nil {
		return nil, err
	}
	return obj.(*metricsprometheusiov1.ResourceRule), err
}

// Update takes the representation of a resourceRule and updates it. Returns the server's representation of the resourceRule, and an error, if there is any.
func (c *FakeResourceRules) Update(resourceRule *metricsprometheusiov1.ResourceRule) (result *metricsprometheusiov1.ResourceRule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(resourcerulesResource, resourceRule), &metricsprometheusiov1.ResourceRule{})
	if obj == nil {
		return nil, err
	}
	return obj.(*metricsprometheusiov1.ResourceRule), err
}

// Delete takes name of the resourceRule and deletes it. Returns an error if one occurs.
func (c *FakeResourceRules) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(resourcerulesResource, name), &metricsprometheusiov1.ResourceRule{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeResourceRules) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(resourcerulesResource, listOptions)

	_, err := c.Fake.Invokes(action, &metricsprometheusiov1.ResourceRuleList{})
	return err
}

// Patch applies the patch and returns the patched resourceRule.
func (c *FakeResourceRules) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *metricsprometheusiov1.ResourceRule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(resourcerulesResource, name, pt, data, subresources...), &metricsprometheusiov1.ResourceRule{})
	if obj == nil {
		return nil, err
	}
	return obj.(*metricsprometheusiov1.ResourceRule), err
}
