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

// DynamicGroupsGetter has a method to return a DynamicGroupInterface.
// A group's client should implement this interface.
type DynamicGroupsGetter interface {
	DynamicGroups(namespace string) DynamicGroupInterface
}

// DynamicGroupInterface has methods to work with DynamicGroup resources.
type DynamicGroupInterface interface {
	Create(*v1alpha1.DynamicGroup) (*v1alpha1.DynamicGroup, error)
	Update(*v1alpha1.DynamicGroup) (*v1alpha1.DynamicGroup, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.DynamicGroup, error)
	List(opts v1.ListOptions) (*v1alpha1.DynamicGroupList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.DynamicGroup, err error)
	DynamicGroupExpansion
}

// dynamicGroups implements DynamicGroupInterface
type dynamicGroups struct {
	client rest.Interface
	ns     string
}

// newDynamicGroups returns a DynamicGroups
func newDynamicGroups(c *OciidentityV1alpha1Client, namespace string) *dynamicGroups {
	return &dynamicGroups{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the dynamicGroup, and returns the corresponding dynamicGroup object, and an error if there is any.
func (c *dynamicGroups) Get(name string, options v1.GetOptions) (result *v1alpha1.DynamicGroup, err error) {
	result = &v1alpha1.DynamicGroup{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("dynamicgroups").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of DynamicGroups that match those selectors.
func (c *dynamicGroups) List(opts v1.ListOptions) (result *v1alpha1.DynamicGroupList, err error) {
	result = &v1alpha1.DynamicGroupList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("dynamicgroups").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested dynamicGroups.
func (c *dynamicGroups) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("dynamicgroups").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a dynamicGroup and creates it.  Returns the server's representation of the dynamicGroup, and an error, if there is any.
func (c *dynamicGroups) Create(dynamicGroup *v1alpha1.DynamicGroup) (result *v1alpha1.DynamicGroup, err error) {
	result = &v1alpha1.DynamicGroup{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("dynamicgroups").
		Body(dynamicGroup).
		Do().
		Into(result)
	return
}

// Update takes the representation of a dynamicGroup and updates it. Returns the server's representation of the dynamicGroup, and an error, if there is any.
func (c *dynamicGroups) Update(dynamicGroup *v1alpha1.DynamicGroup) (result *v1alpha1.DynamicGroup, err error) {
	result = &v1alpha1.DynamicGroup{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("dynamicgroups").
		Name(dynamicGroup.Name).
		Body(dynamicGroup).
		Do().
		Into(result)
	return
}

// Delete takes name of the dynamicGroup and deletes it. Returns an error if one occurs.
func (c *dynamicGroups) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("dynamicgroups").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *dynamicGroups) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("dynamicgroups").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched dynamicGroup.
func (c *dynamicGroups) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.DynamicGroup, err error) {
	result = &v1alpha1.DynamicGroup{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("dynamicgroups").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
