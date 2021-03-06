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

// BackendsGetter has a method to return a BackendInterface.
// A group's client should implement this interface.
type BackendsGetter interface {
	Backends(namespace string) BackendInterface
}

// BackendInterface has methods to work with Backend resources.
type BackendInterface interface {
	Create(*v1alpha1.Backend) (*v1alpha1.Backend, error)
	Update(*v1alpha1.Backend) (*v1alpha1.Backend, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Backend, error)
	List(opts v1.ListOptions) (*v1alpha1.BackendList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Backend, err error)
	BackendExpansion
}

// backends implements BackendInterface
type backends struct {
	client rest.Interface
	ns     string
}

// newBackends returns a Backends
func newBackends(c *OcilbV1alpha1Client, namespace string) *backends {
	return &backends{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the backend, and returns the corresponding backend object, and an error if there is any.
func (c *backends) Get(name string, options v1.GetOptions) (result *v1alpha1.Backend, err error) {
	result = &v1alpha1.Backend{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("backends").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Backends that match those selectors.
func (c *backends) List(opts v1.ListOptions) (result *v1alpha1.BackendList, err error) {
	result = &v1alpha1.BackendList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("backends").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested backends.
func (c *backends) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("backends").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a backend and creates it.  Returns the server's representation of the backend, and an error, if there is any.
func (c *backends) Create(backend *v1alpha1.Backend) (result *v1alpha1.Backend, err error) {
	result = &v1alpha1.Backend{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("backends").
		Body(backend).
		Do().
		Into(result)
	return
}

// Update takes the representation of a backend and updates it. Returns the server's representation of the backend, and an error, if there is any.
func (c *backends) Update(backend *v1alpha1.Backend) (result *v1alpha1.Backend, err error) {
	result = &v1alpha1.Backend{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("backends").
		Name(backend.Name).
		Body(backend).
		Do().
		Into(result)
	return
}

// Delete takes name of the backend and deletes it. Returns an error if one occurs.
func (c *backends) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("backends").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *backends) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("backends").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched backend.
func (c *backends) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Backend, err error) {
	result = &v1alpha1.Backend{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("backends").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
