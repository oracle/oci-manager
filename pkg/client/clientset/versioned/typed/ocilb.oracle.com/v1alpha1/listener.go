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
package v1alpha1

import (
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocilb.oracle.com/v1alpha1"
	scheme "github.com/oracle/oci-manager/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ListenersGetter has a method to return a ListenerInterface.
// A group's client should implement this interface.
type ListenersGetter interface {
	Listeners(namespace string) ListenerInterface
}

// ListenerInterface has methods to work with Listener resources.
type ListenerInterface interface {
	Create(*v1alpha1.Listener) (*v1alpha1.Listener, error)
	Update(*v1alpha1.Listener) (*v1alpha1.Listener, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Listener, error)
	List(opts v1.ListOptions) (*v1alpha1.ListenerList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Listener, err error)
	ListenerExpansion
}

// listeners implements ListenerInterface
type listeners struct {
	client rest.Interface
	ns     string
}

// newListeners returns a Listeners
func newListeners(c *OcilbV1alpha1Client, namespace string) *listeners {
	return &listeners{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the listener, and returns the corresponding listener object, and an error if there is any.
func (c *listeners) Get(name string, options v1.GetOptions) (result *v1alpha1.Listener, err error) {
	result = &v1alpha1.Listener{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("listeners").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Listeners that match those selectors.
func (c *listeners) List(opts v1.ListOptions) (result *v1alpha1.ListenerList, err error) {
	result = &v1alpha1.ListenerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("listeners").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested listeners.
func (c *listeners) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("listeners").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a listener and creates it.  Returns the server's representation of the listener, and an error, if there is any.
func (c *listeners) Create(listener *v1alpha1.Listener) (result *v1alpha1.Listener, err error) {
	result = &v1alpha1.Listener{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("listeners").
		Body(listener).
		Do().
		Into(result)
	return
}

// Update takes the representation of a listener and updates it. Returns the server's representation of the listener, and an error, if there is any.
func (c *listeners) Update(listener *v1alpha1.Listener) (result *v1alpha1.Listener, err error) {
	result = &v1alpha1.Listener{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("listeners").
		Name(listener.Name).
		Body(listener).
		Do().
		Into(result)
	return
}

// Delete takes name of the listener and deletes it. Returns an error if one occurs.
func (c *listeners) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("listeners").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *listeners) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("listeners").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched listener.
func (c *listeners) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Listener, err error) {
	result = &v1alpha1.Listener{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("listeners").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
