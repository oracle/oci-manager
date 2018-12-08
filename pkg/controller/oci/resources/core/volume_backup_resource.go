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

package core

import (
	"context"
	"errors"
	"k8s.io/client-go/kubernetes"
	"os"
	"reflect"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/golang/glog"

	ocisdkcommon "github.com/oracle/oci-go-sdk/common"
	ocicore "github.com/oracle/oci-go-sdk/core"

	ocicommon "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	coregroup "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com"
	ocicorev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	versioned "github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
)

func init() {
	resourcescommon.RegisterResourceTypeWithValidation(
		coregroup.GroupName,
		ocicorev1alpha1.VolumeBackupKind,
		ocicorev1alpha1.VolumeBackupResourcePlural,
		ocicorev1alpha1.VolumeBackupControllerName,
		&ocicorev1alpha1.VolumeBackupValidation,
		NewVolumeBackupAdapter)
}

// VolumeBackupAdapter implements the adapter interface for volume backup resource
type VolumeBackupAdapter struct {
	clientset versioned.Interface
	bsClient  resourcescommon.BlockStorageClientInterface
	ctx       context.Context
}

// NewVolumeBackupAdapter creates a new adapter for volume backup resource
func NewVolumeBackupAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	va := VolumeBackupAdapter{}

	bsClient, err := ocicore.NewBlockstorageClientWithConfigurationProvider(ociconfig)

	if err != nil {
		glog.Errorf("Error creating oci BlockStorage client: %v", err)
		os.Exit(1)
	}

	va.bsClient = &bsClient
	va.clientset = clientset
	va.ctx = context.Background()
	return &va
}

// Kind returns the resource kind string
func (a *VolumeBackupAdapter) Kind() string {
	return ocicorev1alpha1.VolumeBackupKind
}

// Resource returns the plural name of the resource type
func (a *VolumeBackupAdapter) Resource() string {
	return ocicorev1alpha1.VolumeBackupResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *VolumeBackupAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ocicorev1alpha1.SchemeGroupVersion.WithResource(ocicorev1alpha1.VolumeBackupResourcePlural)
}

// ObjectType returns the volume backup type for this adapter
func (a *VolumeBackupAdapter) ObjectType() runtime.Object {
	return &ocicorev1alpha1.VolumeBackup{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *VolumeBackupAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ocicorev1alpha1.VolumeBackup)
	return ok
}

// Copy returns a copy of a volume backup object
func (a *VolumeBackupAdapter) Copy(obj runtime.Object) runtime.Object {
	VolumeBackup := obj.(*ocicorev1alpha1.VolumeBackup)
	return VolumeBackup.DeepCopyObject()
}

// Equivalent checks if two volume backup objects are the same
func (a *VolumeBackupAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	VolumeBackup1 := obj1.(*ocicorev1alpha1.VolumeBackup)
	VolumeBackup2 := obj2.(*ocicorev1alpha1.VolumeBackup)
	if VolumeBackup1.Status.Resource != nil {
		VolumeBackup1.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}

	}

	if VolumeBackup2.Status.Resource != nil {
		VolumeBackup2.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}

	return reflect.DeepEqual(VolumeBackup1, VolumeBackup2)
}

// IsResourceCompliant checks if resource config is complient with CRD spec
func (a *VolumeBackupAdapter) IsResourceCompliant(obj runtime.Object) bool {
	volumeBackup := obj.(*ocicorev1alpha1.VolumeBackup)

	if volumeBackup.Status.Resource == nil {
		return false
	}

	specDisplayName := resourcescommon.Display(volumeBackup.Name, volumeBackup.Spec.DisplayName)

	resource := volumeBackup.Status.Resource
	volumeType := ocicore.VolumeBackupTypeEnum(volumeBackup.Spec.VolumeBackupType)

	if *resource.DisplayName != *specDisplayName ||
		resource.Type != volumeType {
		return false
	}

	return true
}

// IsResourceStatusChanged checks if two vcn objects are the same
func (a *VolumeBackupAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	volumeBackup1 := obj1.(*ocicorev1alpha1.VolumeBackup)
	volumeBackup2 := obj2.(*ocicorev1alpha1.VolumeBackup)

	return volumeBackup1.Status.Resource.LifecycleState != volumeBackup2.Status.Resource.LifecycleState
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *VolumeBackupAdapter) Id(obj runtime.Object) string {
	return obj.(*ocicorev1alpha1.VolumeBackup).GetResourceID()
}

// ObjectMeta returns the object meta struct from the volume backup object
func (a *VolumeBackupAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ocicorev1alpha1.VolumeBackup).ObjectMeta
}

// DependsOn returns a map of volume backup dependencies (objects that the volume backup depends on)
func (a *VolumeBackupAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ocicorev1alpha1.VolumeBackup).Spec.DependsOn
}

// Dependents returns a map of volume backup dependents (objects that depend on the volume backup)
func (a *VolumeBackupAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ocicorev1alpha1.VolumeBackup).Status.Dependents
}

// CreateObject creates the volume backup object
func (a *VolumeBackupAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.VolumeBackup)
	return a.clientset.OcicoreV1alpha1().VolumeBackups(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the volume backup object
func (a *VolumeBackupAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.VolumeBackup)
	return a.clientset.OcicoreV1alpha1().VolumeBackups(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the volume backup object
func (a *VolumeBackupAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var object = obj.(*ocicorev1alpha1.VolumeBackup)
	return a.clientset.OcicoreV1alpha1().VolumeBackups(object.ObjectMeta.Namespace).Delete(object.Name, options)
}

// DependsOnRefs returns the objects that the volume backup depends on
func (a *VolumeBackupAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.VolumeBackup)
	deps := make([]runtime.Object, 0)

	if !resourcescommon.IsOcid(object.Spec.VolumeRef) {
		vol, err := resourcescommon.Volume(a.clientset, object.ObjectMeta.Namespace, object.Spec.VolumeRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, vol)
	}

	return deps, nil
}

// Create creates the volume backup resource in oci
func (a *VolumeBackupAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	var (
		object   = obj.(*ocicorev1alpha1.VolumeBackup)
		volumeId string
	)

	if resourcescommon.IsOcid(object.Spec.VolumeRef) {
		volumeId = object.Spec.VolumeRef
	} else {
		vol, err := resourcescommon.Volume(a.clientset, object.ObjectMeta.Namespace, object.Spec.VolumeRef)
		if err != nil {
			return object, object.Status.HandleError(err)
		}
		if vol.Status.Resource == nil || *vol.Status.Resource.Id == "" {
			return object, object.Status.HandleError(errors.New("Volume resource is not created"))
		}

		volumeId = vol.GetResourceID()
	}

	request := ocicore.CreateVolumeBackupRequest{}
	request.VolumeId = ocisdkcommon.String(volumeId)
	request.DisplayName = resourcescommon.Display(object.Name, object.Spec.DisplayName)
	request.Type = ocicore.CreateVolumeBackupDetailsTypeEnum(object.Spec.VolumeBackupType)
	request.OpcRetryToken = ocisdkcommon.String(string(object.UID))

	r, err := a.bsClient.CreateVolumeBackup(a.ctx, request)

	if err != nil {
		return object, object.Status.HandleError(err)
	}

	return object.SetResource(&r.VolumeBackup), object.Status.HandleError(err)
}

// Delete deletes the volume backup resource in oci
func (a *VolumeBackupAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.VolumeBackup)

	request := ocicore.DeleteVolumeBackupRequest{
		VolumeBackupId: object.Status.Resource.Id,
	}

	_, e := a.bsClient.DeleteVolumeBackup(a.ctx, request)

	if e == nil && object.Status.Resource != nil {
		object.Status.Resource.Id = ocisdkcommon.String("")
	}
	return object, object.Status.HandleError(e)
}

// Get retrieves the volume backup resource from oci
func (a *VolumeBackupAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.VolumeBackup)

	request := ocicore.GetVolumeBackupRequest{
		VolumeBackupId: object.Status.Resource.Id,
	}

	// get VolumeBackup resource
	e := wait.PollImmediate(3*time.Second, 30*time.Second, func() (bool, error) {
		r, e := a.bsClient.GetVolumeBackup(a.ctx, request)
		if e != nil {
			return false, e
		}
		if r.LifecycleState != ocicore.VolumeBackupLifecycleStateCreating {
			object.SetResource(&r.VolumeBackup)
			return true, nil
		}
		return false, e
	})

	return object, object.Status.HandleError(e)
}

// Update updates the volume backup resource in oci
func (a *VolumeBackupAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.VolumeBackup)

	if object.Status.Resource.LifecycleState != ocicore.VolumeBackupLifecycleStateAvailable {
		return object, errors.New(string(object.Status.Resource.LifecycleState))
	}

	request := ocicore.UpdateVolumeBackupRequest{}
	request.VolumeBackupId = object.Status.Resource.Id
	request.DisplayName = resourcescommon.Display(object.Name, object.Spec.DisplayName)

	r, e := a.bsClient.UpdateVolumeBackup(a.ctx, request)

	if e != nil {
		return object, object.Status.HandleError(e)
	}

	return object.SetResource(&r.VolumeBackup), object.Status.HandleError(e)
}

// UpdateForResource calls a common UpdateForResource method to update the volume backup resource in the volume backup object
func (a *VolumeBackupAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
