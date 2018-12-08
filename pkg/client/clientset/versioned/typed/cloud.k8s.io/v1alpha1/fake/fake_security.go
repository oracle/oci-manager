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

// FakeSecurities implements SecurityInterface
type FakeSecurities struct {
	Fake *FakeCloudV1alpha1
	ns   string
}

var securitiesResource = schema.GroupVersionResource{Group: "cloud.k8s.io", Version: "v1alpha1", Resource: "securities"}

var securitiesKind = schema.GroupVersionKind{Group: "cloud.k8s.io", Version: "v1alpha1", Kind: "Security"}

// Get takes name of the security, and returns the corresponding security object, and an error if there is any.
func (c *FakeSecurities) Get(name string, options v1.GetOptions) (result *v1alpha1.Security, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(securitiesResource, c.ns, name), &v1alpha1.Security{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Security), err
}

// List takes label and field selectors, and returns the list of Securities that match those selectors.
func (c *FakeSecurities) List(opts v1.ListOptions) (result *v1alpha1.SecurityList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(securitiesResource, securitiesKind, c.ns, opts), &v1alpha1.SecurityList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.SecurityList{ListMeta: obj.(*v1alpha1.SecurityList).ListMeta}
	for _, item := range obj.(*v1alpha1.SecurityList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested securities.
func (c *FakeSecurities) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(securitiesResource, c.ns, opts))

}

// Create takes the representation of a security and creates it.  Returns the server's representation of the security, and an error, if there is any.
func (c *FakeSecurities) Create(security *v1alpha1.Security) (result *v1alpha1.Security, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(securitiesResource, c.ns, security), &v1alpha1.Security{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Security), err
}

// Update takes the representation of a security and updates it. Returns the server's representation of the security, and an error, if there is any.
func (c *FakeSecurities) Update(security *v1alpha1.Security) (result *v1alpha1.Security, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(securitiesResource, c.ns, security), &v1alpha1.Security{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Security), err
}

// Delete takes name of the security and deletes it. Returns an error if one occurs.
func (c *FakeSecurities) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(securitiesResource, c.ns, name), &v1alpha1.Security{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSecurities) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(securitiesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.SecurityList{})
	return err
}

// Patch applies the patch and returns the patched security.
func (c *FakeSecurities) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Security, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(securitiesResource, c.ns, name, data, subresources...), &v1alpha1.Security{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Security), err
}
