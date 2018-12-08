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
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	scheme "github.com/oracle/oci-manager/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// VcnsGetter has a method to return a VcnInterface.
// A group's client should implement this interface.
type VcnsGetter interface {
	Vcns(namespace string) VcnInterface
}

// VcnInterface has methods to work with Vcn resources.
type VcnInterface interface {
	Create(*v1alpha1.Vcn) (*v1alpha1.Vcn, error)
	Update(*v1alpha1.Vcn) (*v1alpha1.Vcn, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Vcn, error)
	List(opts v1.ListOptions) (*v1alpha1.VcnList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Vcn, err error)
	VcnExpansion
}

// vcns implements VcnInterface
type vcns struct {
	client rest.Interface
	ns     string
}

// newVcns returns a Vcns
func newVcns(c *OcicoreV1alpha1Client, namespace string) *vcns {
	return &vcns{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the vcn, and returns the corresponding vcn object, and an error if there is any.
func (c *vcns) Get(name string, options v1.GetOptions) (result *v1alpha1.Vcn, err error) {
	result = &v1alpha1.Vcn{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("vcns").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Vcns that match those selectors.
func (c *vcns) List(opts v1.ListOptions) (result *v1alpha1.VcnList, err error) {
	result = &v1alpha1.VcnList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("vcns").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested vcns.
func (c *vcns) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("vcns").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a vcn and creates it.  Returns the server's representation of the vcn, and an error, if there is any.
func (c *vcns) Create(vcn *v1alpha1.Vcn) (result *v1alpha1.Vcn, err error) {
	result = &v1alpha1.Vcn{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("vcns").
		Body(vcn).
		Do().
		Into(result)
	return
}

// Update takes the representation of a vcn and updates it. Returns the server's representation of the vcn, and an error, if there is any.
func (c *vcns) Update(vcn *v1alpha1.Vcn) (result *v1alpha1.Vcn, err error) {
	result = &v1alpha1.Vcn{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("vcns").
		Name(vcn.Name).
		Body(vcn).
		Do().
		Into(result)
	return
}

// Delete takes name of the vcn and deletes it. Returns an error if one occurs.
func (c *vcns) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("vcns").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *vcns) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("vcns").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched vcn.
func (c *vcns) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Vcn, err error) {
	result = &v1alpha1.Vcn{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("vcns").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
