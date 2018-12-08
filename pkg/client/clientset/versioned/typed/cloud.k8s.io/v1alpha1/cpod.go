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

// CpodsGetter has a method to return a CpodInterface.
// A group's client should implement this interface.
type CpodsGetter interface {
	Cpods(namespace string) CpodInterface
}

// CpodInterface has methods to work with Cpod resources.
type CpodInterface interface {
	Create(*v1alpha1.Cpod) (*v1alpha1.Cpod, error)
	Update(*v1alpha1.Cpod) (*v1alpha1.Cpod, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Cpod, error)
	List(opts v1.ListOptions) (*v1alpha1.CpodList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Cpod, err error)
	CpodExpansion
}

// cpods implements CpodInterface
type cpods struct {
	client rest.Interface
	ns     string
}

// newCpods returns a Cpods
func newCpods(c *CloudV1alpha1Client, namespace string) *cpods {
	return &cpods{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the cpod, and returns the corresponding cpod object, and an error if there is any.
func (c *cpods) Get(name string, options v1.GetOptions) (result *v1alpha1.Cpod, err error) {
	result = &v1alpha1.Cpod{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cpods").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Cpods that match those selectors.
func (c *cpods) List(opts v1.ListOptions) (result *v1alpha1.CpodList, err error) {
	result = &v1alpha1.CpodList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cpods").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested cpods.
func (c *cpods) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("cpods").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a cpod and creates it.  Returns the server's representation of the cpod, and an error, if there is any.
func (c *cpods) Create(cpod *v1alpha1.Cpod) (result *v1alpha1.Cpod, err error) {
	result = &v1alpha1.Cpod{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("cpods").
		Body(cpod).
		Do().
		Into(result)
	return
}

// Update takes the representation of a cpod and updates it. Returns the server's representation of the cpod, and an error, if there is any.
func (c *cpods) Update(cpod *v1alpha1.Cpod) (result *v1alpha1.Cpod, err error) {
	result = &v1alpha1.Cpod{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("cpods").
		Name(cpod.Name).
		Body(cpod).
		Do().
		Into(result)
	return
}

// Delete takes name of the cpod and deletes it. Returns an error if one occurs.
func (c *cpods) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cpods").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *cpods) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cpods").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched cpod.
func (c *cpods) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Cpod, err error) {
	result = &v1alpha1.Cpod{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("cpods").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
