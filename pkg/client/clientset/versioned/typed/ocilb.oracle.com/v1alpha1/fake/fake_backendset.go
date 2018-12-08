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
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocilb.oracle.com/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeBackendSets implements BackendSetInterface
type FakeBackendSets struct {
	Fake *FakeOcilbV1alpha1
	ns   string
}

var backendsetsResource = schema.GroupVersionResource{Group: "ocilb.oracle.com", Version: "v1alpha1", Resource: "backendsets"}

var backendsetsKind = schema.GroupVersionKind{Group: "ocilb.oracle.com", Version: "v1alpha1", Kind: "BackendSet"}

// Get takes name of the backendSet, and returns the corresponding backendSet object, and an error if there is any.
func (c *FakeBackendSets) Get(name string, options v1.GetOptions) (result *v1alpha1.BackendSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(backendsetsResource, c.ns, name), &v1alpha1.BackendSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BackendSet), err
}

// List takes label and field selectors, and returns the list of BackendSets that match those selectors.
func (c *FakeBackendSets) List(opts v1.ListOptions) (result *v1alpha1.BackendSetList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(backendsetsResource, backendsetsKind, c.ns, opts), &v1alpha1.BackendSetList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.BackendSetList{ListMeta: obj.(*v1alpha1.BackendSetList).ListMeta}
	for _, item := range obj.(*v1alpha1.BackendSetList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested backendSets.
func (c *FakeBackendSets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(backendsetsResource, c.ns, opts))

}

// Create takes the representation of a backendSet and creates it.  Returns the server's representation of the backendSet, and an error, if there is any.
func (c *FakeBackendSets) Create(backendSet *v1alpha1.BackendSet) (result *v1alpha1.BackendSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(backendsetsResource, c.ns, backendSet), &v1alpha1.BackendSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BackendSet), err
}

// Update takes the representation of a backendSet and updates it. Returns the server's representation of the backendSet, and an error, if there is any.
func (c *FakeBackendSets) Update(backendSet *v1alpha1.BackendSet) (result *v1alpha1.BackendSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(backendsetsResource, c.ns, backendSet), &v1alpha1.BackendSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BackendSet), err
}

// Delete takes name of the backendSet and deletes it. Returns an error if one occurs.
func (c *FakeBackendSets) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(backendsetsResource, c.ns, name), &v1alpha1.BackendSet{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeBackendSets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(backendsetsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.BackendSetList{})
	return err
}

// Patch applies the patch and returns the patched backendSet.
func (c *FakeBackendSets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.BackendSet, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(backendsetsResource, c.ns, name, data, subresources...), &v1alpha1.BackendSet{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.BackendSet), err
}
