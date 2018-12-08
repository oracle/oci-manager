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
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocidb.oracle.com/v1alpha1"
	scheme "github.com/oracle/oci-manager/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// AutonomousDatabasesGetter has a method to return a AutonomousDatabaseInterface.
// A group's client should implement this interface.
type AutonomousDatabasesGetter interface {
	AutonomousDatabases(namespace string) AutonomousDatabaseInterface
}

// AutonomousDatabaseInterface has methods to work with AutonomousDatabase resources.
type AutonomousDatabaseInterface interface {
	Create(*v1alpha1.AutonomousDatabase) (*v1alpha1.AutonomousDatabase, error)
	Update(*v1alpha1.AutonomousDatabase) (*v1alpha1.AutonomousDatabase, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.AutonomousDatabase, error)
	List(opts v1.ListOptions) (*v1alpha1.AutonomousDatabaseList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.AutonomousDatabase, err error)
	AutonomousDatabaseExpansion
}

// autonomousDatabases implements AutonomousDatabaseInterface
type autonomousDatabases struct {
	client rest.Interface
	ns     string
}

// newAutonomousDatabases returns a AutonomousDatabases
func newAutonomousDatabases(c *OcidbV1alpha1Client, namespace string) *autonomousDatabases {
	return &autonomousDatabases{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the autonomousDatabase, and returns the corresponding autonomousDatabase object, and an error if there is any.
func (c *autonomousDatabases) Get(name string, options v1.GetOptions) (result *v1alpha1.AutonomousDatabase, err error) {
	result = &v1alpha1.AutonomousDatabase{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("autonomousdatabases").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of AutonomousDatabases that match those selectors.
func (c *autonomousDatabases) List(opts v1.ListOptions) (result *v1alpha1.AutonomousDatabaseList, err error) {
	result = &v1alpha1.AutonomousDatabaseList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("autonomousdatabases").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested autonomousDatabases.
func (c *autonomousDatabases) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("autonomousdatabases").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a autonomousDatabase and creates it.  Returns the server's representation of the autonomousDatabase, and an error, if there is any.
func (c *autonomousDatabases) Create(autonomousDatabase *v1alpha1.AutonomousDatabase) (result *v1alpha1.AutonomousDatabase, err error) {
	result = &v1alpha1.AutonomousDatabase{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("autonomousdatabases").
		Body(autonomousDatabase).
		Do().
		Into(result)
	return
}

// Update takes the representation of a autonomousDatabase and updates it. Returns the server's representation of the autonomousDatabase, and an error, if there is any.
func (c *autonomousDatabases) Update(autonomousDatabase *v1alpha1.AutonomousDatabase) (result *v1alpha1.AutonomousDatabase, err error) {
	result = &v1alpha1.AutonomousDatabase{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("autonomousdatabases").
		Name(autonomousDatabase.Name).
		Body(autonomousDatabase).
		Do().
		Into(result)
	return
}

// Delete takes name of the autonomousDatabase and deletes it. Returns an error if one occurs.
func (c *autonomousDatabases) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("autonomousdatabases").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *autonomousDatabases) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("autonomousdatabases").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched autonomousDatabase.
func (c *autonomousDatabases) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.AutonomousDatabase, err error) {
	result = &v1alpha1.AutonomousDatabase{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("autonomousdatabases").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
