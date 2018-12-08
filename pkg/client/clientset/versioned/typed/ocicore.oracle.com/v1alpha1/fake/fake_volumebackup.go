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
package fake

import (
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeVolumeBackups implements VolumeBackupInterface
type FakeVolumeBackups struct {
	Fake *FakeOcicoreV1alpha1
	ns   string
}

var volumebackupsResource = schema.GroupVersionResource{Group: "ocicore.oracle.com", Version: "v1alpha1", Resource: "volumebackups"}

var volumebackupsKind = schema.GroupVersionKind{Group: "ocicore.oracle.com", Version: "v1alpha1", Kind: "VolumeBackup"}

// Get takes name of the volumeBackup, and returns the corresponding volumeBackup object, and an error if there is any.
func (c *FakeVolumeBackups) Get(name string, options v1.GetOptions) (result *v1alpha1.VolumeBackup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(volumebackupsResource, c.ns, name), &v1alpha1.VolumeBackup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VolumeBackup), err
}

// List takes label and field selectors, and returns the list of VolumeBackups that match those selectors.
func (c *FakeVolumeBackups) List(opts v1.ListOptions) (result *v1alpha1.VolumeBackupList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(volumebackupsResource, volumebackupsKind, c.ns, opts), &v1alpha1.VolumeBackupList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.VolumeBackupList{ListMeta: obj.(*v1alpha1.VolumeBackupList).ListMeta}
	for _, item := range obj.(*v1alpha1.VolumeBackupList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested volumeBackups.
func (c *FakeVolumeBackups) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(volumebackupsResource, c.ns, opts))

}

// Create takes the representation of a volumeBackup and creates it.  Returns the server's representation of the volumeBackup, and an error, if there is any.
func (c *FakeVolumeBackups) Create(volumeBackup *v1alpha1.VolumeBackup) (result *v1alpha1.VolumeBackup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(volumebackupsResource, c.ns, volumeBackup), &v1alpha1.VolumeBackup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VolumeBackup), err
}

// Update takes the representation of a volumeBackup and updates it. Returns the server's representation of the volumeBackup, and an error, if there is any.
func (c *FakeVolumeBackups) Update(volumeBackup *v1alpha1.VolumeBackup) (result *v1alpha1.VolumeBackup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(volumebackupsResource, c.ns, volumeBackup), &v1alpha1.VolumeBackup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VolumeBackup), err
}

// Delete takes name of the volumeBackup and deletes it. Returns an error if one occurs.
func (c *FakeVolumeBackups) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(volumebackupsResource, c.ns, name), &v1alpha1.VolumeBackup{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVolumeBackups) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(volumebackupsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.VolumeBackupList{})
	return err
}

// Patch applies the patch and returns the patched volumeBackup.
func (c *FakeVolumeBackups) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VolumeBackup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(volumebackupsResource, c.ns, name, data, subresources...), &v1alpha1.VolumeBackup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VolumeBackup), err
}
