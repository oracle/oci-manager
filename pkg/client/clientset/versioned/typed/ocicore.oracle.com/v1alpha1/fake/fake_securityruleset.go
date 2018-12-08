/*
Copyright 2018 Oracle and/or its affiliates. All rights reserved.

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
package fake

import (
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSecurityRuleSets implements SecurityRuleSetInterface
type FakeSecurityRuleSets struct {
	Fake *FakeOcicoreV1alpha1
	ns   string
}

var securityrulesetsResource = schema.GroupVersionResource{Group: "ocicore.oracle.com", Version: "v1alpha1", Resource: "securityrulesets"}

var securityrulesetsKind = schema.GroupVersionKind{Group: "ocicore.oracle.com", Version: "v1alpha1", Kind: "SecurityRuleSet"}

// Get takes name of the securityRuleSet, and returns the corresponding securityRuleSet object, and an error if there is any.
func (c *FakeSecurityRuleSets) Get(name string, options v1.GetOptions) (result *v1alpha1.SecurityRuleSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(securityrulesetsResource, c.ns, name), &v1alpha1.SecurityRuleSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SecurityRuleSet), err
}

// List takes label and field selectors, and returns the list of SecurityRuleSets that match those selectors.
func (c *FakeSecurityRuleSets) List(opts v1.ListOptions) (result *v1alpha1.SecurityRuleSetList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(securityrulesetsResource, securityrulesetsKind, c.ns, opts), &v1alpha1.SecurityRuleSetList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.SecurityRuleSetList{ListMeta: obj.(*v1alpha1.SecurityRuleSetList).ListMeta}
	for _, item := range obj.(*v1alpha1.SecurityRuleSetList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested securityRuleSets.
func (c *FakeSecurityRuleSets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(securityrulesetsResource, c.ns, opts))

}

// Create takes the representation of a securityRuleSet and creates it.  Returns the server's representation of the securityRuleSet, and an error, if there is any.
func (c *FakeSecurityRuleSets) Create(securityRuleSet *v1alpha1.SecurityRuleSet) (result *v1alpha1.SecurityRuleSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(securityrulesetsResource, c.ns, securityRuleSet), &v1alpha1.SecurityRuleSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SecurityRuleSet), err
}

// Update takes the representation of a securityRuleSet and updates it. Returns the server's representation of the securityRuleSet, and an error, if there is any.
func (c *FakeSecurityRuleSets) Update(securityRuleSet *v1alpha1.SecurityRuleSet) (result *v1alpha1.SecurityRuleSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(securityrulesetsResource, c.ns, securityRuleSet), &v1alpha1.SecurityRuleSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SecurityRuleSet), err
}

// Delete takes name of the securityRuleSet and deletes it. Returns an error if one occurs.
func (c *FakeSecurityRuleSets) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(securityrulesetsResource, c.ns, name), &v1alpha1.SecurityRuleSet{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSecurityRuleSets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(securityrulesetsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.SecurityRuleSetList{})
	return err
}

// Patch applies the patch and returns the patched securityRuleSet.
func (c *FakeSecurityRuleSets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.SecurityRuleSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(securityrulesetsResource, c.ns, name, data, subresources...), &v1alpha1.SecurityRuleSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SecurityRuleSet), err
}
