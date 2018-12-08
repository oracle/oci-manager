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

// BackendSetsGetter has a method to return a BackendSetInterface.
// A group's client should implement this interface.
type BackendSetsGetter interface {
	BackendSets(namespace string) BackendSetInterface
}

// BackendSetInterface has methods to work with BackendSet resources.
type BackendSetInterface interface {
	Create(*v1alpha1.BackendSet) (*v1alpha1.BackendSet, error)
	Update(*v1alpha1.BackendSet) (*v1alpha1.BackendSet, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.BackendSet, error)
	List(opts v1.ListOptions) (*v1alpha1.BackendSetList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.BackendSet, err error)
	BackendSetExpansion
}

// backendSets implements BackendSetInterface
type backendSets struct {
	client rest.Interface
	ns     string
}

// newBackendSets returns a BackendSets
func newBackendSets(c *OcilbV1alpha1Client, namespace string) *backendSets {
	return &backendSets{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the backendSet, and returns the corresponding backendSet object, and an error if there is any.
func (c *backendSets) Get(name string, options v1.GetOptions) (result *v1alpha1.BackendSet, err error) {
	result = &v1alpha1.BackendSet{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("backendsets").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of BackendSets that match those selectors.
func (c *backendSets) List(opts v1.ListOptions) (result *v1alpha1.BackendSetList, err error) {
	result = &v1alpha1.BackendSetList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("backendsets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested backendSets.
func (c *backendSets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("backendsets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a backendSet and creates it.  Returns the server's representation of the backendSet, and an error, if there is any.
func (c *backendSets) Create(backendSet *v1alpha1.BackendSet) (result *v1alpha1.BackendSet, err error) {
	result = &v1alpha1.BackendSet{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("backendsets").
		Body(backendSet).
		Do().
		Into(result)
	return
}

// Update takes the representation of a backendSet and updates it. Returns the server's representation of the backendSet, and an error, if there is any.
func (c *backendSets) Update(backendSet *v1alpha1.BackendSet) (result *v1alpha1.BackendSet, err error) {
	result = &v1alpha1.BackendSet{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("backendsets").
		Name(backendSet.Name).
		Body(backendSet).
		Do().
		Into(result)
	return
}

// Delete takes name of the backendSet and deletes it. Returns an error if one occurs.
func (c *backendSets) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("backendsets").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *backendSets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("backendsets").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched backendSet.
func (c *backendSets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.BackendSet, err error) {
	result = &v1alpha1.BackendSet{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("backendsets").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
