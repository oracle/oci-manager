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

// DhcpOptionsGetter has a method to return a DhcpOptionInterface.
// A group's client should implement this interface.
type DhcpOptionsGetter interface {
	DhcpOptions(namespace string) DhcpOptionInterface
}

// DhcpOptionInterface has methods to work with DhcpOption resources.
type DhcpOptionInterface interface {
	Create(*v1alpha1.DhcpOption) (*v1alpha1.DhcpOption, error)
	Update(*v1alpha1.DhcpOption) (*v1alpha1.DhcpOption, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.DhcpOption, error)
	List(opts v1.ListOptions) (*v1alpha1.DhcpOptionList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.DhcpOption, err error)
	DhcpOptionExpansion
}

// dhcpOptions implements DhcpOptionInterface
type dhcpOptions struct {
	client rest.Interface
	ns     string
}

// newDhcpOptions returns a DhcpOptions
func newDhcpOptions(c *OcicoreV1alpha1Client, namespace string) *dhcpOptions {
	return &dhcpOptions{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the dhcpOption, and returns the corresponding dhcpOption object, and an error if there is any.
func (c *dhcpOptions) Get(name string, options v1.GetOptions) (result *v1alpha1.DhcpOption, err error) {
	result = &v1alpha1.DhcpOption{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("dhcpoptions").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of DhcpOptions that match those selectors.
func (c *dhcpOptions) List(opts v1.ListOptions) (result *v1alpha1.DhcpOptionList, err error) {
	result = &v1alpha1.DhcpOptionList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("dhcpoptions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested dhcpOptions.
func (c *dhcpOptions) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("dhcpoptions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a dhcpOption and creates it.  Returns the server's representation of the dhcpOption, and an error, if there is any.
func (c *dhcpOptions) Create(dhcpOption *v1alpha1.DhcpOption) (result *v1alpha1.DhcpOption, err error) {
	result = &v1alpha1.DhcpOption{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("dhcpoptions").
		Body(dhcpOption).
		Do().
		Into(result)
	return
}

// Update takes the representation of a dhcpOption and updates it. Returns the server's representation of the dhcpOption, and an error, if there is any.
func (c *dhcpOptions) Update(dhcpOption *v1alpha1.DhcpOption) (result *v1alpha1.DhcpOption, err error) {
	result = &v1alpha1.DhcpOption{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("dhcpoptions").
		Name(dhcpOption.Name).
		Body(dhcpOption).
		Do().
		Into(result)
	return
}

// Delete takes name of the dhcpOption and deletes it. Returns an error if one occurs.
func (c *dhcpOptions) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("dhcpoptions").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *dhcpOptions) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("dhcpoptions").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched dhcpOption.
func (c *dhcpOptions) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.DhcpOption, err error) {
	result = &v1alpha1.DhcpOption{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("dhcpoptions").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
