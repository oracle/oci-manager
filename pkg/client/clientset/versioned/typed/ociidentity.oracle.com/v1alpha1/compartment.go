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
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com/v1alpha1"
	scheme "github.com/oracle/oci-manager/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// CompartmentsGetter has a method to return a CompartmentInterface.
// A group's client should implement this interface.
type CompartmentsGetter interface {
	Compartments(namespace string) CompartmentInterface
}

// CompartmentInterface has methods to work with Compartment resources.
type CompartmentInterface interface {
	Create(*v1alpha1.Compartment) (*v1alpha1.Compartment, error)
	Update(*v1alpha1.Compartment) (*v1alpha1.Compartment, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Compartment, error)
	List(opts v1.ListOptions) (*v1alpha1.CompartmentList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Compartment, err error)
	CompartmentExpansion
}

// compartments implements CompartmentInterface
type compartments struct {
	client rest.Interface
	ns     string
}

// newCompartments returns a Compartments
func newCompartments(c *OciidentityV1alpha1Client, namespace string) *compartments {
	return &compartments{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the compartment, and returns the corresponding compartment object, and an error if there is any.
func (c *compartments) Get(name string, options v1.GetOptions) (result *v1alpha1.Compartment, err error) {
	result = &v1alpha1.Compartment{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("compartments").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Compartments that match those selectors.
func (c *compartments) List(opts v1.ListOptions) (result *v1alpha1.CompartmentList, err error) {
	result = &v1alpha1.CompartmentList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("compartments").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested compartments.
func (c *compartments) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("compartments").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a compartment and creates it.  Returns the server's representation of the compartment, and an error, if there is any.
func (c *compartments) Create(compartment *v1alpha1.Compartment) (result *v1alpha1.Compartment, err error) {
	result = &v1alpha1.Compartment{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("compartments").
		Body(compartment).
		Do().
		Into(result)
	return
}

// Update takes the representation of a compartment and updates it. Returns the server's representation of the compartment, and an error, if there is any.
func (c *compartments) Update(compartment *v1alpha1.Compartment) (result *v1alpha1.Compartment, err error) {
	result = &v1alpha1.Compartment{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("compartments").
		Name(compartment.Name).
		Body(compartment).
		Do().
		Into(result)
	return
}

// Delete takes name of the compartment and deletes it. Returns an error if one occurs.
func (c *compartments) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("compartments").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *compartments) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("compartments").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched compartment.
func (c *compartments) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Compartment, err error) {
	result = &v1alpha1.Compartment{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("compartments").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
