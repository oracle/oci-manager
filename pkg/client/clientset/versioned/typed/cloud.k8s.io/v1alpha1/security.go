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

// SecuritiesGetter has a method to return a SecurityInterface.
// A group's client should implement this interface.
type SecuritiesGetter interface {
	Securities(namespace string) SecurityInterface
}

// SecurityInterface has methods to work with Security resources.
type SecurityInterface interface {
	Create(*v1alpha1.Security) (*v1alpha1.Security, error)
	Update(*v1alpha1.Security) (*v1alpha1.Security, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Security, error)
	List(opts v1.ListOptions) (*v1alpha1.SecurityList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Security, err error)
	SecurityExpansion
}

// securities implements SecurityInterface
type securities struct {
	client rest.Interface
	ns     string
}

// newSecurities returns a Securities
func newSecurities(c *CloudV1alpha1Client, namespace string) *securities {
	return &securities{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the security, and returns the corresponding security object, and an error if there is any.
func (c *securities) Get(name string, options v1.GetOptions) (result *v1alpha1.Security, err error) {
	result = &v1alpha1.Security{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("securities").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Securities that match those selectors.
func (c *securities) List(opts v1.ListOptions) (result *v1alpha1.SecurityList, err error) {
	result = &v1alpha1.SecurityList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("securities").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested securities.
func (c *securities) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("securities").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a security and creates it.  Returns the server's representation of the security, and an error, if there is any.
func (c *securities) Create(security *v1alpha1.Security) (result *v1alpha1.Security, err error) {
	result = &v1alpha1.Security{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("securities").
		Body(security).
		Do().
		Into(result)
	return
}

// Update takes the representation of a security and updates it. Returns the server's representation of the security, and an error, if there is any.
func (c *securities) Update(security *v1alpha1.Security) (result *v1alpha1.Security, err error) {
	result = &v1alpha1.Security{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("securities").
		Name(security.Name).
		Body(security).
		Do().
		Into(result)
	return
}

// Delete takes name of the security and deletes it. Returns an error if one occurs.
func (c *securities) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("securities").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *securities) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("securities").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched security.
func (c *securities) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Security, err error) {
	result = &v1alpha1.Security{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("securities").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
