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
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/cloud.k8s.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeComputes implements ComputeInterface
type FakeComputes struct {
	Fake *FakeCloudV1alpha1
	ns   string
}

var computesResource = schema.GroupVersionResource{Group: "cloud.k8s.io", Version: "v1alpha1", Resource: "computes"}

var computesKind = schema.GroupVersionKind{Group: "cloud.k8s.io", Version: "v1alpha1", Kind: "Compute"}

// Get takes name of the compute, and returns the corresponding compute object, and an error if there is any.
func (c *FakeComputes) Get(name string, options v1.GetOptions) (result *v1alpha1.Compute, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(computesResource, c.ns, name), &v1alpha1.Compute{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Compute), err
}

// List takes label and field selectors, and returns the list of Computes that match those selectors.
func (c *FakeComputes) List(opts v1.ListOptions) (result *v1alpha1.ComputeList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(computesResource, computesKind, c.ns, opts), &v1alpha1.ComputeList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ComputeList{ListMeta: obj.(*v1alpha1.ComputeList).ListMeta}
	for _, item := range obj.(*v1alpha1.ComputeList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested computes.
func (c *FakeComputes) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(computesResource, c.ns, opts))

}

// Create takes the representation of a compute and creates it.  Returns the server's representation of the compute, and an error, if there is any.
func (c *FakeComputes) Create(compute *v1alpha1.Compute) (result *v1alpha1.Compute, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(computesResource, c.ns, compute), &v1alpha1.Compute{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Compute), err
}

// Update takes the representation of a compute and updates it. Returns the server's representation of the compute, and an error, if there is any.
func (c *FakeComputes) Update(compute *v1alpha1.Compute) (result *v1alpha1.Compute, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(computesResource, c.ns, compute), &v1alpha1.Compute{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Compute), err
}

// Delete takes name of the compute and deletes it. Returns an error if one occurs.
func (c *FakeComputes) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(computesResource, c.ns, name), &v1alpha1.Compute{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeComputes) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(computesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ComputeList{})
	return err
}

// Patch applies the patch and returns the patched compute.
func (c *FakeComputes) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Compute, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(computesResource, c.ns, name, data, subresources...), &v1alpha1.Compute{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Compute), err
}
