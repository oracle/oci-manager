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

// InternetGatewaiesGetter has a method to return a InternetGatewayInterface.
// A group's client should implement this interface.
type InternetGatewaiesGetter interface {
	InternetGatewaies(namespace string) InternetGatewayInterface
}

// InternetGatewayInterface has methods to work with InternetGateway resources.
type InternetGatewayInterface interface {
	Create(*v1alpha1.InternetGateway) (*v1alpha1.InternetGateway, error)
	Update(*v1alpha1.InternetGateway) (*v1alpha1.InternetGateway, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.InternetGateway, error)
	List(opts v1.ListOptions) (*v1alpha1.InternetGatewayList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.InternetGateway, err error)
	InternetGatewayExpansion
}

// internetGatewaies implements InternetGatewayInterface
type internetGatewaies struct {
	client rest.Interface
	ns     string
}

// newInternetGatewaies returns a InternetGatewaies
func newInternetGatewaies(c *OcicoreV1alpha1Client, namespace string) *internetGatewaies {
	return &internetGatewaies{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the internetGateway, and returns the corresponding internetGateway object, and an error if there is any.
func (c *internetGatewaies) Get(name string, options v1.GetOptions) (result *v1alpha1.InternetGateway, err error) {
	result = &v1alpha1.InternetGateway{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("internetgatewaies").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of InternetGatewaies that match those selectors.
func (c *internetGatewaies) List(opts v1.ListOptions) (result *v1alpha1.InternetGatewayList, err error) {
	result = &v1alpha1.InternetGatewayList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("internetgatewaies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested internetGatewaies.
func (c *internetGatewaies) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("internetgatewaies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a internetGateway and creates it.  Returns the server's representation of the internetGateway, and an error, if there is any.
func (c *internetGatewaies) Create(internetGateway *v1alpha1.InternetGateway) (result *v1alpha1.InternetGateway, err error) {
	result = &v1alpha1.InternetGateway{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("internetgatewaies").
		Body(internetGateway).
		Do().
		Into(result)
	return
}

// Update takes the representation of a internetGateway and updates it. Returns the server's representation of the internetGateway, and an error, if there is any.
func (c *internetGatewaies) Update(internetGateway *v1alpha1.InternetGateway) (result *v1alpha1.InternetGateway, err error) {
	result = &v1alpha1.InternetGateway{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("internetgatewaies").
		Name(internetGateway.Name).
		Body(internetGateway).
		Do().
		Into(result)
	return
}

// Delete takes name of the internetGateway and deletes it. Returns an error if one occurs.
func (c *internetGatewaies) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("internetgatewaies").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *internetGatewaies) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("internetgatewaies").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched internetGateway.
func (c *internetGatewaies) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.InternetGateway, err error) {
	result = &v1alpha1.InternetGateway{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("internetgatewaies").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
