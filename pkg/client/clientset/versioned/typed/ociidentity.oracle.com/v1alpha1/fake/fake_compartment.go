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

// FakeCompartments implements CompartmentInterface
type FakeCompartments struct {
	Fake *FakeOciidentityV1alpha1
	ns   string
}

var compartmentsResource = schema.GroupVersionResource{Group: "ociidentity.oracle.com", Version: "v1alpha1", Resource: "compartments"}

var compartmentsKind = schema.GroupVersionKind{Group: "ociidentity.oracle.com", Version: "v1alpha1", Kind: "Compartment"}

// Get takes name of the compartment, and returns the corresponding compartment object, and an error if there is any.
func (c *FakeCompartments) Get(name string, options v1.GetOptions) (result *v1alpha1.Compartment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(compartmentsResource, c.ns, name), &v1alpha1.Compartment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Compartment), err
}

// List takes label and field selectors, and returns the list of Compartments that match those selectors.
func (c *FakeCompartments) List(opts v1.ListOptions) (result *v1alpha1.CompartmentList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(compartmentsResource, compartmentsKind, c.ns, opts), &v1alpha1.CompartmentList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CompartmentList{ListMeta: obj.(*v1alpha1.CompartmentList).ListMeta}
	for _, item := range obj.(*v1alpha1.CompartmentList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested compartments.
func (c *FakeCompartments) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(compartmentsResource, c.ns, opts))

}

// Create takes the representation of a compartment and creates it.  Returns the server's representation of the compartment, and an error, if there is any.
func (c *FakeCompartments) Create(compartment *v1alpha1.Compartment) (result *v1alpha1.Compartment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(compartmentsResource, c.ns, compartment), &v1alpha1.Compartment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Compartment), err
}

// Update takes the representation of a compartment and updates it. Returns the server's representation of the compartment, and an error, if there is any.
func (c *FakeCompartments) Update(compartment *v1alpha1.Compartment) (result *v1alpha1.Compartment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(compartmentsResource, c.ns, compartment), &v1alpha1.Compartment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Compartment), err
}

// Delete takes name of the compartment and deletes it. Returns an error if one occurs.
func (c *FakeCompartments) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(compartmentsResource, c.ns, name), &v1alpha1.Compartment{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCompartments) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(compartmentsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.CompartmentList{})
	return err
}

// Patch applies the patch and returns the patched compartment.
func (c *FakeCompartments) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Compartment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(compartmentsResource, c.ns, name, data, subresources...), &v1alpha1.Compartment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Compartment), err
}
