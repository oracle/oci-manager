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

// FakeCpods implements CpodInterface
type FakeCpods struct {
	Fake *FakeCloudV1alpha1
	ns   string
}

var cpodsResource = schema.GroupVersionResource{Group: "cloud.k8s.io", Version: "v1alpha1", Resource: "cpods"}

var cpodsKind = schema.GroupVersionKind{Group: "cloud.k8s.io", Version: "v1alpha1", Kind: "Cpod"}

// Get takes name of the cpod, and returns the corresponding cpod object, and an error if there is any.
func (c *FakeCpods) Get(name string, options v1.GetOptions) (result *v1alpha1.Cpod, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(cpodsResource, c.ns, name), &v1alpha1.Cpod{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Cpod), err
}

// List takes label and field selectors, and returns the list of Cpods that match those selectors.
func (c *FakeCpods) List(opts v1.ListOptions) (result *v1alpha1.CpodList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(cpodsResource, cpodsKind, c.ns, opts), &v1alpha1.CpodList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CpodList{ListMeta: obj.(*v1alpha1.CpodList).ListMeta}
	for _, item := range obj.(*v1alpha1.CpodList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested cpods.
func (c *FakeCpods) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(cpodsResource, c.ns, opts))

}

// Create takes the representation of a cpod and creates it.  Returns the server's representation of the cpod, and an error, if there is any.
func (c *FakeCpods) Create(cpod *v1alpha1.Cpod) (result *v1alpha1.Cpod, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(cpodsResource, c.ns, cpod), &v1alpha1.Cpod{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Cpod), err
}

// Update takes the representation of a cpod and updates it. Returns the server's representation of the cpod, and an error, if there is any.
func (c *FakeCpods) Update(cpod *v1alpha1.Cpod) (result *v1alpha1.Cpod, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(cpodsResource, c.ns, cpod), &v1alpha1.Cpod{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Cpod), err
}

// Delete takes name of the cpod and deletes it. Returns an error if one occurs.
func (c *FakeCpods) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(cpodsResource, c.ns, name), &v1alpha1.Cpod{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCpods) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(cpodsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.CpodList{})
	return err
}

// Patch applies the patch and returns the patched cpod.
func (c *FakeCpods) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Cpod, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(cpodsResource, c.ns, name, data, subresources...), &v1alpha1.Cpod{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Cpod), err
}
