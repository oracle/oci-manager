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

// FakeListeners implements ListenerInterface
type FakeListeners struct {
	Fake *FakeOcilbV1alpha1
	ns   string
}

var listenersResource = schema.GroupVersionResource{Group: "ocilb.oracle.com", Version: "v1alpha1", Resource: "listeners"}

var listenersKind = schema.GroupVersionKind{Group: "ocilb.oracle.com", Version: "v1alpha1", Kind: "Listener"}

// Get takes name of the listener, and returns the corresponding listener object, and an error if there is any.
func (c *FakeListeners) Get(name string, options v1.GetOptions) (result *v1alpha1.Listener, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(listenersResource, c.ns, name), &v1alpha1.Listener{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Listener), err
}

// List takes label and field selectors, and returns the list of Listeners that match those selectors.
func (c *FakeListeners) List(opts v1.ListOptions) (result *v1alpha1.ListenerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(listenersResource, listenersKind, c.ns, opts), &v1alpha1.ListenerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ListenerList{ListMeta: obj.(*v1alpha1.ListenerList).ListMeta}
	for _, item := range obj.(*v1alpha1.ListenerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested listeners.
func (c *FakeListeners) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(listenersResource, c.ns, opts))

}

// Create takes the representation of a listener and creates it.  Returns the server's representation of the listener, and an error, if there is any.
func (c *FakeListeners) Create(listener *v1alpha1.Listener) (result *v1alpha1.Listener, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(listenersResource, c.ns, listener), &v1alpha1.Listener{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Listener), err
}

// Update takes the representation of a listener and updates it. Returns the server's representation of the listener, and an error, if there is any.
func (c *FakeListeners) Update(listener *v1alpha1.Listener) (result *v1alpha1.Listener, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(listenersResource, c.ns, listener), &v1alpha1.Listener{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Listener), err
}

// Delete takes name of the listener and deletes it. Returns an error if one occurs.
func (c *FakeListeners) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(listenersResource, c.ns, name), &v1alpha1.Listener{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeListeners) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(listenersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ListenerList{})
	return err
}

// Patch applies the patch and returns the patched listener.
func (c *FakeListeners) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Listener, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(listenersResource, c.ns, name, data, subresources...), &v1alpha1.Listener{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Listener), err
}
