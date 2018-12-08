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

	"fmt"
	ocicommon "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	coregroup "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com"
	ocicorev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	versioned "github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
)

func init() {
	resourcescommon.RegisterResourceTypeWithValidation(
		coregroup.GroupName,
		ocicorev1alpha1.VolumeKind,
		ocicorev1alpha1.VolumeResourcePlural,
		ocicorev1alpha1.VolumeControllerName,
		&ocicorev1alpha1.VolumeValidation,
		NewVolumeAdapter)
}

// VolumeAdapter implements the adapter interface for volume resource
type VolumeAdapter struct {
	clientset versioned.Interface
	bsClient  resourcescommon.BlockStorageClientInterface
	cClient   resourcescommon.ComputeClientInterface
	ctx       context.Context
}

// NewVolumeAdapter creates a new adapter for volume resource
func NewVolumeAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	va := VolumeAdapter{}

	cClient, err := ocicore.NewComputeClientWithConfigurationProvider(ociconfig)

	if err != nil {
		glog.Errorf("Error creating oci Compute client: %v", err)
		os.Exit(1)
	}

	bsClient, err := ocicore.NewBlockstorageClientWithConfigurationProvider(ociconfig)

	if err != nil {
		glog.Errorf("Error creating oci BlockStorage client: %v", err)
		os.Exit(1)
	}

	va.cClient = &cClient
	va.bsClient = &bsClient
	va.clientset = clientset
	va.ctx = context.Background()
	return &va
}

// Kind returns the resource kind string
func (a *VolumeAdapter) Kind() string {
	return ocicorev1alpha1.VolumeKind
}

// Resource returns the plural name of the resource type
func (a *VolumeAdapter) Resource() string {
	return ocicorev1alpha1.VolumeResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *VolumeAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ocicorev1alpha1.SchemeGroupVersion.WithResource(ocicorev1alpha1.VolumeResourcePlural)
}

// ObjectType returns the volume type for this adapter
func (a *VolumeAdapter) ObjectType() runtime.Object {
	return &ocicorev1alpha1.Volume{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *VolumeAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ocicorev1alpha1.Volume)
	return ok
}

// Copy returns a copy of a volume object
func (a *VolumeAdapter) Copy(obj runtime.Object) runtime.Object {
	volume := obj.(*ocicorev1alpha1.Volume)
	return volume.DeepCopyObject()
}

// Equivalent checks if two volume objects are the same
func (a *VolumeAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	volume1 := obj1.(*ocicorev1alpha1.Volume)
	volume2 := obj2.(*ocicorev1alpha1.Volume)
	if volume1.Status.Resource != nil {
		volume1.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}

	}

	if volume1.Status.Attachment != nil {
		timeCreated := volume1.Status.Attachment.GetTimeCreated()
		*timeCreated = ocisdkcommon.SDKTime{}
	}

	if volume2.Status.Resource != nil {
		volume2.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}

	if volume2.Status.Attachment != nil {
		timeCreated := volume2.Status.Attachment.GetTimeCreated()
		*timeCreated = ocisdkcommon.SDKTime{}
	}

	if volume1.Status.Resource != nil && volume2.Status.Resource != nil {
		volume1.Status.Resource.SourceDetails = nil
		volume2.Status.Resource.SourceDetails = nil
	}

	return reflect.DeepEqual(volume1, volume2)
}

// IsResourceCompliant checks if resource config is complient with CRD spec
func (a *VolumeAdapter) IsResourceCompliant(obj runtime.Object) bool {

	volume := obj.(*ocicorev1alpha1.Volume)

	if volume.Status.Resource == nil {
		return false
	}

	resource := volume.Status.Resource

	if resource.LifecycleState == ocicore.VolumeLifecycleStateProvisioning ||
		resource.LifecycleState == ocicore.VolumeLifecycleStateTerminating {
		return true
	}

	if resource.LifecycleState == ocicore.VolumeLifecycleStateTerminated {
		return false
	}

	specDisplayName := resourcescommon.Display(volume.Name, volume.Spec.DisplayName)

	if *resource.DisplayName != *specDisplayName ||
		*resource.AvailabilityDomain != volume.Spec.AvailabilityDomain ||
		*resource.SizeInGBs != *ocisdkcommon.Int64(volume.Spec.SizeInGBs) {
		return false
	}

	if volume.Spec.InstanceRef != "" {
		if volume.Status.Attachment == nil {
			return false
		}

		if volume.Status.Attachment.GetLifecycleState() == ocicore.VolumeAttachmentLifecycleStateDetached {
			return false
		}

	}

	if volume.Status.Attachment != nil &&
		volume.Spec.AttachmentType != volume.Status.Attachment.AttachmentType {
		return false
	}

	return true
}

// IsResourceStatusChanged checks if two vcn objects are the same
func (a *VolumeAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	volume1 := obj1.(*ocicorev1alpha1.Volume)
	volume2 := obj2.(*ocicorev1alpha1.Volume)

	if volume1.Status.Resource.LifecycleState != volume2.Status.Resource.LifecycleState {
		return true
	}

	if volume1.Status.AttachmentState != volume2.Status.AttachmentState {
		return true
	}

	return false
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *VolumeAdapter) Id(obj runtime.Object) string {
	return obj.(*ocicorev1alpha1.Volume).GetResourceID()
}

// ObjectMeta returns the object meta struct from the volume object
func (a *VolumeAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ocicorev1alpha1.Volume).ObjectMeta
}

// DependsOn returns a map of volume dependencies (objects that the volume depends on)
func (a *VolumeAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ocicorev1alpha1.Volume).Spec.DependsOn
}

// Dependents returns a map of volume dependents (objects that depend on the volume)
func (a *VolumeAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ocicorev1alpha1.Volume).Status.Dependents
}

// CreateObject creates the volume object
func (a *VolumeAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Volume)
	return a.clientset.OcicoreV1alpha1().Volumes(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the volume object
func (a *VolumeAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Volume)
	return a.clientset.OcicoreV1alpha1().Volumes(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the volume object
func (a *VolumeAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var object = obj.(*ocicorev1alpha1.Volume)
	return a.clientset.OcicoreV1alpha1().Volumes(object.ObjectMeta.Namespace).Delete(object.Name, options)
}

// DependsOnRefs returns the objects that the volume depends on
func (a *VolumeAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Volume)
	deps := make([]runtime.Object, 0)

	if !resourcescommon.IsOcid(object.Spec.CompartmentRef) {
		compartment, err := resourcescommon.Compartment(a.clientset, object.ObjectMeta.Namespace, object.Spec.CompartmentRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, compartment)
	}
	return deps, nil
}

// Create creates the volume resource in oci
func (a *VolumeAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	var (
		object        = obj.(*ocicorev1alpha1.Volume)
		compartmentId string
	)

	if resourcescommon.IsOcid(object.Spec.CompartmentRef) {
		compartmentId = object.Spec.CompartmentRef
	} else {
		compartment, err := resourcescommon.Compartment(a.clientset, object.ObjectMeta.Namespace, object.Spec.CompartmentRef)
		if err != nil {
			return object, object.Status.HandleError(err)
		}
		if !compartment.IsResource() {
			return object, object.Status.HandleError(errors.New("Compartment resource does not exist"))
		}
		compartmentId = compartment.GetResourceID()
	}

	request := ocicore.CreateVolumeRequest{}
	request.CompartmentId = ocisdkcommon.String(compartmentId)
	request.DisplayName = resourcescommon.Display(object.Name, object.Spec.DisplayName)
	request.AvailabilityDomain = ocisdkcommon.String(object.Spec.AvailabilityDomain)
	request.SizeInGBs = ocisdkcommon.Int64(object.Spec.SizeInGBs)

	request.OpcRetryToken = ocisdkcommon.String(string(object.UID))

	r, err := a.bsClient.CreateVolume(a.ctx, request)

	if err != nil {
		return object, object.Status.HandleError(err)
	}

	return object.SetResource(&r.Volume), object.Status.HandleError(err)
}

// Delete deletes the volume resource in oci
func (a *VolumeAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Volume)

	if object.Status.Attachment != nil && object.Status.Attachment.GetId() != nil {

		attrequest := ocicore.GetVolumeAttachmentRequest{
			VolumeAttachmentId: object.Status.Attachment.GetId(),
		}

		attresp, e := a.cClient.GetVolumeAttachment(a.ctx, attrequest)

		if e != nil {
			return object, object.Status.HandleError(e)
		}

		if attresp.GetLifecycleState() == ocicore.VolumeAttachmentLifecycleStateAttached {
			detach := ocicore.DetachVolumeRequest{
				VolumeAttachmentId: object.Status.Attachment.GetId(),
			}

			_, e := a.cClient.DetachVolume(a.ctx, detach)
			if e != nil {
				return object, object.Status.HandleError(e)
			}

			attrequest := ocicore.GetVolumeAttachmentRequest{
				VolumeAttachmentId: object.Status.Attachment.GetId(),
			}

			getresp, e := a.cClient.GetVolumeAttachment(a.ctx, attrequest)
			if getresp.GetLifecycleState() != ocicore.VolumeAttachmentLifecycleStateDetached {
				return object, fmt.Errorf("Volume %s is still attached", object.Name)
			}
		} else if attresp.GetLifecycleState() == ocicore.VolumeAttachmentLifecycleStateDetached {
			object.SetAttachment("", nil)
			object.Status.AttachmentState = ocicore.VolumeAttachmentLifecycleStateDetached
		} else {
			return object, fmt.Errorf("Volume %s is still attached", object.Name)
		}
	}

	request := ocicore.DeleteVolumeRequest{
		VolumeId: object.Status.Resource.Id,
	}

	_, e := a.bsClient.DeleteVolume(a.ctx, request)

	if e == nil && object.Status.Resource != nil {
		object.Status.Resource.Id = ocisdkcommon.String("")
	}
	return object, object.Status.HandleError(e)
}

// Get retrieves the volume resource from oci
func (a *VolumeAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Volume)

	request := ocicore.GetVolumeRequest{
		VolumeId: object.Status.Resource.Id,
	}

	e := wait.PollImmediate(3*time.Second, 30*time.Second, func() (bool, error) {
		r, e := a.bsClient.GetVolume(a.ctx, request)
		if e != nil {
			return false, e
		}
		if r.LifecycleState != ocicore.VolumeLifecycleStateProvisioning {
			object.SetResource(&r.Volume)
			return true, nil
		}
		return false, e
	})
	if object.Spec.InstanceRef == "" || object.Status.Attachment == nil {
		object.Status.AttachmentState = ocicore.VolumeAttachmentLifecycleStateDetached
		return object, object.Status.HandleError(e)
	}

	attrequest := ocicore.GetVolumeAttachmentRequest{
		VolumeAttachmentId: object.Status.Attachment.GetId(),
	}

	attresp, e := a.cClient.GetVolumeAttachment(a.ctx, attrequest)

	if e != nil {
		return object, object.Status.HandleError(e)
	}

	object.SetAttachment(object.Spec.AttachmentType, &attresp.VolumeAttachment)
	object.Status.AttachmentState = attresp.VolumeAttachment.GetLifecycleState()

	return object, object.Status.HandleError(e)

}

// Update updates the volume resource in oci
func (a *VolumeAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Volume)

	if object.Status.Resource.LifecycleState != ocicore.VolumeLifecycleStateAvailable {
		return object, errors.New(string(object.Status.Resource.LifecycleState))
	}

	request := ocicore.UpdateVolumeRequest{}
	request.VolumeId = object.Status.Resource.Id
	request.DisplayName = resourcescommon.Display(object.Name, object.Spec.DisplayName)

	r, e := a.bsClient.UpdateVolume(a.ctx, request)

	if e != nil {
		return object, object.Status.HandleError(e)
	}

	if object.Spec.InstanceRef == "" {
		return object.SetResource(&r.Volume), object.Status.HandleError(e)
	}

	object.SetResource(&r.Volume)

	// volume is not attached, lets create attachment
	if object.Status.Attachment == nil || object.Status.AttachmentState == ocicore.VolumeAttachmentLifecycleStateDetached {

		instance, err := a.clientset.OcicoreV1alpha1().Instances(object.ObjectMeta.Namespace).Get(object.Spec.InstanceRef, metav1.GetOptions{})
		if err != nil {
			return object, object.Status.HandleError(err)
		}
		if !instance.IsResource() {
			return object, object.Status.HandleError(errors.New("Instance resource does not exist or not in ready state"))
		}

		var attrequest ocicore.AttachVolumeRequest

		dispName := *resourcescommon.Display(object.Name, object.Spec.DisplayName) + "-attachment"

		if object.Spec.AttachmentType == "iscsi" {
			attrequest = ocicore.AttachVolumeRequest{
				AttachVolumeDetails: ocicore.AttachIScsiVolumeDetails{
					InstanceId:  instance.Status.Resource.Id,
					VolumeId:    object.Status.Resource.Id,
					DisplayName: &dispName,
					IsReadOnly:  ocisdkcommon.Bool(false),
				},
			}
		} else {
			attrequest = ocicore.AttachVolumeRequest{
				AttachVolumeDetails: ocicore.AttachParavirtualizedVolumeDetails{
					InstanceId:  instance.Status.Resource.Id,
					VolumeId:    object.Status.Resource.Id,
					DisplayName: &dispName,
					IsReadOnly:  ocisdkcommon.Bool(false),
				},
			}
		}

		attresp, e := a.cClient.AttachVolume(a.ctx, attrequest)

		if e != nil {
			return object, object.Status.HandleError(e)
		}

		object.SetAttachment(object.Spec.AttachmentType, &attresp.VolumeAttachment)
		object.Status.AttachmentState = attresp.VolumeAttachment.GetLifecycleState()

	} else {

		attrequest := ocicore.GetVolumeAttachmentRequest{
			VolumeAttachmentId: object.Status.Attachment.GetId(),
		}
		attresp, e := a.cClient.GetVolumeAttachment(a.ctx, attrequest)

		if e != nil {
			return object, object.Status.HandleError(e)
		}

		object.SetAttachment(object.Spec.AttachmentType, &attresp.VolumeAttachment)
		object.Status.AttachmentState = attresp.VolumeAttachment.GetLifecycleState()

	}
	return object, object.Status.HandleError(e)
}

// UpdateForResource calls a common UpdateForResource method to update the volume resource in the volume object
func (a *VolumeAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
