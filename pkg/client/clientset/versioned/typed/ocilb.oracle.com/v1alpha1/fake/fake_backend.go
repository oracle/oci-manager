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

// FakeBackends implements BackendInterface
type FakeBackends struct {
	Fake *FakeOcilbV1alpha1
	ns   string
}

var backendsResource = schema.GroupVersionResource{Group: "ocilb.oracle.com", Version: "v1alpha1", Resource: "backends"}

var backendsKind = schema.GroupVersionKind{Group: "ocilb.oracle.com", Version: "v1alpha1", Kind: "Backend"}

// Get takes name of the backend, and returns the corresponding backend object, and an error if there is any.
func (c *FakeBackends) Get(name string, options v1.GetOptions) (result *v1alpha1.Backend, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(backendsResource, c.ns, name), &v1alpha1.Backend{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Backend), err
}

// List takes label and field selectors, and returns the list of Backends that match those selectors.
func (c *FakeBackends) List(opts v1.ListOptions) (result *v1alpha1.BackendList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(backendsResource, backendsKind, c.ns, opts), &v1alpha1.BackendList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.BackendList{ListMeta: obj.(*v1alpha1.BackendList).ListMeta}
	for _, item := range obj.(*v1alpha1.BackendList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested backends.
func (c *FakeBackends) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(backendsResource, c.ns, opts))

}

// Create takes the representation of a backend and creates it.  Returns the server's representation of the backend, and an error, if there is any.
func (c *FakeBackends) Create(backend *v1alpha1.Backend) (result *v1alpha1.Backend, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(backendsResource, c.ns, backend), &v1alpha1.Backend{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Backend), err
}

// Update takes the representation of a backend and updates it. Returns the server's representation of the backend, and an error, if there is any.
func (c *FakeBackends) Update(backend *v1alpha1.Backend) (result *v1alpha1.Backend, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(backendsResource, c.ns, backend), &v1alpha1.Backend{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Backend), err
}

// Delete takes name of the backend and deletes it. Returns an error if one occurs.
func (c *FakeBackends) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(backendsResource, c.ns, name), &v1alpha1.Backend{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeBackends) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(backendsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.BackendList{})
	return err
}

// Patch applies the patch and returns the patched backend.
func (c *FakeBackends) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Backend, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(backendsResource, c.ns, name, data, subresources...), &v1alpha1.Backend{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Backend), err
}
