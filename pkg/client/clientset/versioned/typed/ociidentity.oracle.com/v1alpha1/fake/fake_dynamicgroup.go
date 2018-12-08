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
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeDynamicGroups implements DynamicGroupInterface
type FakeDynamicGroups struct {
	Fake *FakeOciidentityV1alpha1
	ns   string
}

var dynamicgroupsResource = schema.GroupVersionResource{Group: "ociidentity.oracle.com", Version: "v1alpha1", Resource: "dynamicgroups"}

var dynamicgroupsKind = schema.GroupVersionKind{Group: "ociidentity.oracle.com", Version: "v1alpha1", Kind: "DynamicGroup"}

// Get takes name of the dynamicGroup, and returns the corresponding dynamicGroup object, and an error if there is any.
func (c *FakeDynamicGroups) Get(name string, options v1.GetOptions) (result *v1alpha1.DynamicGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(dynamicgroupsResource, c.ns, name), &v1alpha1.DynamicGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DynamicGroup), err
}

// List takes label and field selectors, and returns the list of DynamicGroups that match those selectors.
func (c *FakeDynamicGroups) List(opts v1.ListOptions) (result *v1alpha1.DynamicGroupList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(dynamicgroupsResource, dynamicgroupsKind, c.ns, opts), &v1alpha1.DynamicGroupList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.DynamicGroupList{ListMeta: obj.(*v1alpha1.DynamicGroupList).ListMeta}
	for _, item := range obj.(*v1alpha1.DynamicGroupList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested dynamicGroups.
func (c *FakeDynamicGroups) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(dynamicgroupsResource, c.ns, opts))

}

// Create takes the representation of a dynamicGroup and creates it.  Returns the server's representation of the dynamicGroup, and an error, if there is any.
func (c *FakeDynamicGroups) Create(dynamicGroup *v1alpha1.DynamicGroup) (result *v1alpha1.DynamicGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(dynamicgroupsResource, c.ns, dynamicGroup), &v1alpha1.DynamicGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DynamicGroup), err
}

// Update takes the representation of a dynamicGroup and updates it. Returns the server's representation of the dynamicGroup, and an error, if there is any.
func (c *FakeDynamicGroups) Update(dynamicGroup *v1alpha1.DynamicGroup) (result *v1alpha1.DynamicGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(dynamicgroupsResource, c.ns, dynamicGroup), &v1alpha1.DynamicGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DynamicGroup), err
}

// Delete takes name of the dynamicGroup and deletes it. Returns an error if one occurs.
func (c *FakeDynamicGroups) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(dynamicgroupsResource, c.ns, name), &v1alpha1.DynamicGroup{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDynamicGroups) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(dynamicgroupsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.DynamicGroupList{})
	return err
}

// Patch applies the patch and returns the patched dynamicGroup.
func (c *FakeDynamicGroups) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.DynamicGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(dynamicgroupsResource, c.ns, name, data, subresources...), &v1alpha1.DynamicGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DynamicGroup), err
}
