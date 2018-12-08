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

// VolumeBackupsGetter has a method to return a VolumeBackupInterface.
// A group's client should implement this interface.
type VolumeBackupsGetter interface {
	VolumeBackups(namespace string) VolumeBackupInterface
}

// VolumeBackupInterface has methods to work with VolumeBackup resources.
type VolumeBackupInterface interface {
	Create(*v1alpha1.VolumeBackup) (*v1alpha1.VolumeBackup, error)
	Update(*v1alpha1.VolumeBackup) (*v1alpha1.VolumeBackup, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.VolumeBackup, error)
	List(opts v1.ListOptions) (*v1alpha1.VolumeBackupList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VolumeBackup, err error)
	VolumeBackupExpansion
}

// volumeBackups implements VolumeBackupInterface
type volumeBackups struct {
	client rest.Interface
	ns     string
}

// newVolumeBackups returns a VolumeBackups
func newVolumeBackups(c *OcicoreV1alpha1Client, namespace string) *volumeBackups {
	return &volumeBackups{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the volumeBackup, and returns the corresponding volumeBackup object, and an error if there is any.
func (c *volumeBackups) Get(name string, options v1.GetOptions) (result *v1alpha1.VolumeBackup, err error) {
	result = &v1alpha1.VolumeBackup{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("volumebackups").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of VolumeBackups that match those selectors.
func (c *volumeBackups) List(opts v1.ListOptions) (result *v1alpha1.VolumeBackupList, err error) {
	result = &v1alpha1.VolumeBackupList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("volumebackups").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested volumeBackups.
func (c *volumeBackups) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("volumebackups").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a volumeBackup and creates it.  Returns the server's representation of the volumeBackup, and an error, if there is any.
func (c *volumeBackups) Create(volumeBackup *v1alpha1.VolumeBackup) (result *v1alpha1.VolumeBackup, err error) {
	result = &v1alpha1.VolumeBackup{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("volumebackups").
		Body(volumeBackup).
		Do().
		Into(result)
	return
}

// Update takes the representation of a volumeBackup and updates it. Returns the server's representation of the volumeBackup, and an error, if there is any.
func (c *volumeBackups) Update(volumeBackup *v1alpha1.VolumeBackup) (result *v1alpha1.VolumeBackup, err error) {
	result = &v1alpha1.VolumeBackup{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("volumebackups").
		Name(volumeBackup.Name).
		Body(volumeBackup).
		Do().
		Into(result)
	return
}

// Delete takes name of the volumeBackup and deletes it. Returns an error if one occurs.
func (c *volumeBackups) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("volumebackups").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *volumeBackups) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("volumebackups").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched volumeBackup.
func (c *volumeBackups) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VolumeBackup, err error) {
	result = &v1alpha1.VolumeBackup{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("volumebackups").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
