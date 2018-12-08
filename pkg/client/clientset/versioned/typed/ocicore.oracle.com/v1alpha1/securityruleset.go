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

// SecurityRuleSetsGetter has a method to return a SecurityRuleSetInterface.
// A group's client should implement this interface.
type SecurityRuleSetsGetter interface {
	SecurityRuleSets(namespace string) SecurityRuleSetInterface
}

// SecurityRuleSetInterface has methods to work with SecurityRuleSet resources.
type SecurityRuleSetInterface interface {
	Create(*v1alpha1.SecurityRuleSet) (*v1alpha1.SecurityRuleSet, error)
	Update(*v1alpha1.SecurityRuleSet) (*v1alpha1.SecurityRuleSet, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.SecurityRuleSet, error)
	List(opts v1.ListOptions) (*v1alpha1.SecurityRuleSetList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.SecurityRuleSet, err error)
	SecurityRuleSetExpansion
}

// securityRuleSets implements SecurityRuleSetInterface
type securityRuleSets struct {
	client rest.Interface
	ns     string
}

// newSecurityRuleSets returns a SecurityRuleSets
func newSecurityRuleSets(c *OcicoreV1alpha1Client, namespace string) *securityRuleSets {
	return &securityRuleSets{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the securityRuleSet, and returns the corresponding securityRuleSet object, and an error if there is any.
func (c *securityRuleSets) Get(name string, options v1.GetOptions) (result *v1alpha1.SecurityRuleSet, err error) {
	result = &v1alpha1.SecurityRuleSet{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("securityrulesets").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SecurityRuleSets that match those selectors.
func (c *securityRuleSets) List(opts v1.ListOptions) (result *v1alpha1.SecurityRuleSetList, err error) {
	result = &v1alpha1.SecurityRuleSetList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("securityrulesets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested securityRuleSets.
func (c *securityRuleSets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("securityrulesets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a securityRuleSet and creates it.  Returns the server's representation of the securityRuleSet, and an error, if there is any.
func (c *securityRuleSets) Create(securityRuleSet *v1alpha1.SecurityRuleSet) (result *v1alpha1.SecurityRuleSet, err error) {
	result = &v1alpha1.SecurityRuleSet{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("securityrulesets").
		Body(securityRuleSet).
		Do().
		Into(result)
	return
}

// Update takes the representation of a securityRuleSet and updates it. Returns the server's representation of the securityRuleSet, and an error, if there is any.
func (c *securityRuleSets) Update(securityRuleSet *v1alpha1.SecurityRuleSet) (result *v1alpha1.SecurityRuleSet, err error) {
	result = &v1alpha1.SecurityRuleSet{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("securityrulesets").
		Name(securityRuleSet.Name).
		Body(securityRuleSet).
		Do().
		Into(result)
	return
}

// Delete takes name of the securityRuleSet and deletes it. Returns an error if one occurs.
func (c *securityRuleSets) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("securityrulesets").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *securityRuleSets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("securityrulesets").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched securityRuleSet.
func (c *securityRuleSets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.SecurityRuleSet, err error) {
	result = &v1alpha1.SecurityRuleSet{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("securityrulesets").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
