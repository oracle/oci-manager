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

// RouteTablesGetter has a method to return a RouteTableInterface.
// A group's client should implement this interface.
type RouteTablesGetter interface {
	RouteTables(namespace string) RouteTableInterface
}

// RouteTableInterface has methods to work with RouteTable resources.
type RouteTableInterface interface {
	Create(*v1alpha1.RouteTable) (*v1alpha1.RouteTable, error)
	Update(*v1alpha1.RouteTable) (*v1alpha1.RouteTable, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.RouteTable, error)
	List(opts v1.ListOptions) (*v1alpha1.RouteTableList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.RouteTable, err error)
	RouteTableExpansion
}

// routeTables implements RouteTableInterface
type routeTables struct {
	client rest.Interface
	ns     string
}

// newRouteTables returns a RouteTables
func newRouteTables(c *OcicoreV1alpha1Client, namespace string) *routeTables {
	return &routeTables{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the routeTable, and returns the corresponding routeTable object, and an error if there is any.
func (c *routeTables) Get(name string, options v1.GetOptions) (result *v1alpha1.RouteTable, err error) {
	result = &v1alpha1.RouteTable{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("routetables").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of RouteTables that match those selectors.
func (c *routeTables) List(opts v1.ListOptions) (result *v1alpha1.RouteTableList, err error) {
	result = &v1alpha1.RouteTableList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("routetables").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested routeTables.
func (c *routeTables) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("routetables").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a routeTable and creates it.  Returns the server's representation of the routeTable, and an error, if there is any.
func (c *routeTables) Create(routeTable *v1alpha1.RouteTable) (result *v1alpha1.RouteTable, err error) {
	result = &v1alpha1.RouteTable{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("routetables").
		Body(routeTable).
		Do().
		Into(result)
	return
}

// Update takes the representation of a routeTable and updates it. Returns the server's representation of the routeTable, and an error, if there is any.
func (c *routeTables) Update(routeTable *v1alpha1.RouteTable) (result *v1alpha1.RouteTable, err error) {
	result = &v1alpha1.RouteTable{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("routetables").
		Name(routeTable.Name).
		Body(routeTable).
		Do().
		Into(result)
	return
}

// Delete takes name of the routeTable and deletes it. Returns an error if one occurs.
func (c *routeTables) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("routetables").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *routeTables) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("routetables").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched routeTable.
func (c *routeTables) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.RouteTable, err error) {
	result = &v1alpha1.RouteTable{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("routetables").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
