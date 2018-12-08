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
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/cloud.k8s.io/v1alpha1"
	scheme "github.com/oracle/oci-manager/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ComputesGetter has a method to return a ComputeInterface.
// A group's client should implement this interface.
type ComputesGetter interface {
	Computes(namespace string) ComputeInterface
}

// ComputeInterface has methods to work with Compute resources.
type ComputeInterface interface {
	Create(*v1alpha1.Compute) (*v1alpha1.Compute, error)
	Update(*v1alpha1.Compute) (*v1alpha1.Compute, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Compute, error)
	List(opts v1.ListOptions) (*v1alpha1.ComputeList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Compute, err error)
	ComputeExpansion
}

// computes implements ComputeInterface
type computes struct {
	client rest.Interface
	ns     string
}

// newComputes returns a Computes
func newComputes(c *CloudV1alpha1Client, namespace string) *computes {
	return &computes{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the compute, and returns the corresponding compute object, and an error if there is any.
func (c *computes) Get(name string, options v1.GetOptions) (result *v1alpha1.Compute, err error) {
	result = &v1alpha1.Compute{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("computes").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Computes that match those selectors.
func (c *computes) List(opts v1.ListOptions) (result *v1alpha1.ComputeList, err error) {
	result = &v1alpha1.ComputeList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("computes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested computes.
func (c *computes) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("computes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a compute and creates it.  Returns the server's representation of the compute, and an error, if there is any.
func (c *computes) Create(compute *v1alpha1.Compute) (result *v1alpha1.Compute, err error) {
	result = &v1alpha1.Compute{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("computes").
		Body(compute).
		Do().
		Into(result)
	return
}

// Update takes the representation of a compute and updates it. Returns the server's representation of the compute, and an error, if there is any.
func (c *computes) Update(compute *v1alpha1.Compute) (result *v1alpha1.Compute, err error) {
	result = &v1alpha1.Compute{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("computes").
		Name(compute.Name).
		Body(compute).
		Do().
		Into(result)
	return
}

// Delete takes name of the compute and deletes it. Returns an error if one occurs.
func (c *computes) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("computes").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *computes) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("computes").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched compute.
func (c *computes) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Compute, err error) {
	result = &v1alpha1.Compute{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("computes").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
