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

// FakeDhcpOptions implements DhcpOptionInterface
type FakeDhcpOptions struct {
	Fake *FakeOcicoreV1alpha1
	ns   string
}

var dhcpoptionsResource = schema.GroupVersionResource{Group: "ocicore.oracle.com", Version: "v1alpha1", Resource: "dhcpoptions"}

var dhcpoptionsKind = schema.GroupVersionKind{Group: "ocicore.oracle.com", Version: "v1alpha1", Kind: "DhcpOption"}

// Get takes name of the dhcpOption, and returns the corresponding dhcpOption object, and an error if there is any.
func (c *FakeDhcpOptions) Get(name string, options v1.GetOptions) (result *v1alpha1.DhcpOption, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(dhcpoptionsResource, c.ns, name), &v1alpha1.DhcpOption{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DhcpOption), err
}

// List takes label and field selectors, and returns the list of DhcpOptions that match those selectors.
func (c *FakeDhcpOptions) List(opts v1.ListOptions) (result *v1alpha1.DhcpOptionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(dhcpoptionsResource, dhcpoptionsKind, c.ns, opts), &v1alpha1.DhcpOptionList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.DhcpOptionList{ListMeta: obj.(*v1alpha1.DhcpOptionList).ListMeta}
	for _, item := range obj.(*v1alpha1.DhcpOptionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested dhcpOptions.
func (c *FakeDhcpOptions) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(dhcpoptionsResource, c.ns, opts))

}

// Create takes the representation of a dhcpOption and creates it.  Returns the server's representation of the dhcpOption, and an error, if there is any.
func (c *FakeDhcpOptions) Create(dhcpOption *v1alpha1.DhcpOption) (result *v1alpha1.DhcpOption, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(dhcpoptionsResource, c.ns, dhcpOption), &v1alpha1.DhcpOption{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DhcpOption), err
}

// Update takes the representation of a dhcpOption and updates it. Returns the server's representation of the dhcpOption, and an error, if there is any.
func (c *FakeDhcpOptions) Update(dhcpOption *v1alpha1.DhcpOption) (result *v1alpha1.DhcpOption, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(dhcpoptionsResource, c.ns, dhcpOption), &v1alpha1.DhcpOption{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DhcpOption), err
}

// Delete takes name of the dhcpOption and deletes it. Returns an error if one occurs.
func (c *FakeDhcpOptions) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(dhcpoptionsResource, c.ns, name), &v1alpha1.DhcpOption{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDhcpOptions) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(dhcpoptionsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.DhcpOptionList{})
	return err
}

// Patch applies the patch and returns the patched dhcpOption.
func (c *FakeDhcpOptions) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.DhcpOption, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(dhcpoptionsResource, c.ns, name, data, subresources...), &v1alpha1.DhcpOption{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DhcpOption), err
}
