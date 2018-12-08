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
	"errors"
	ocicommon "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	coregroup "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com"
	ocicorev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"reflect"

	ocisdkcommon "github.com/oracle/oci-go-sdk/common"
	ocicore "github.com/oracle/oci-go-sdk/core"

	"context"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strconv"
)

// Adapter functions
func init() {
	resourcescommon.RegisterResourceTypeWithValidation(
		coregroup.GroupName,
		ocicorev1alpha1.InstanceKind,
		ocicorev1alpha1.InstanceResourcePlural,
		ocicorev1alpha1.InstanceControllerName,
		&ocicorev1alpha1.InstanceValidation,
		NewInstanceAdapter)
}

// InstanceAdapter implements the adapter interface for instance resource
type InstanceAdapter struct {
	clientset versioned.Interface
	ctx       context.Context
	cClient   resourcescommon.ComputeClientInterface
	bsClient  resourcescommon.BlockStorageClientInterface
	vcnClient resourcescommon.VcnClientInterface
}

// NewInstanceAdapter creates a new adapter for instance resource
func NewInstanceAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	ia := InstanceAdapter{}

	cClient, err := ocicore.NewComputeClientWithConfigurationProvider(ociconfig)

	if err != nil {
		glog.Errorf("Error creating oci Compute client: %v", err)
		os.Exit(1)
	}

	vcnClient, err := ocicore.NewVirtualNetworkClientWithConfigurationProvider(ociconfig)

	if err != nil {
		glog.Errorf("Error creating oci VCN client: %v", err)
		os.Exit(1)
	}

	bsClient, err := ocicore.NewBlockstorageClientWithConfigurationProvider(ociconfig)

	if err != nil {
		glog.Errorf("Error creating oci BlockStorage client: %v", err)
		os.Exit(1)
	}

	ia.cClient = &cClient
	ia.vcnClient = &vcnClient
	ia.bsClient = &bsClient
	ia.clientset = clientset
	ia.ctx = context.Background()
	return &ia
}

// Kind returns the resource kind string
func (a *InstanceAdapter) Kind() string {
	return ocicorev1alpha1.InstanceKind
}

// Resource returns the plural name of the resource type
func (a *InstanceAdapter) Resource() string {
	return ocicorev1alpha1.InstanceResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *InstanceAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ocicorev1alpha1.SchemeGroupVersion.WithResource(ocicorev1alpha1.InstanceResourcePlural)
}

// ObjectType returns the instance type for this adapter
func (a *InstanceAdapter) ObjectType() runtime.Object {
	return &ocicorev1alpha1.Instance{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *InstanceAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ocicorev1alpha1.Instance)
	return ok
}

// Copy returns a copy of a instance object
func (a *InstanceAdapter) Copy(obj runtime.Object) runtime.Object {
	instance := obj.(*ocicorev1alpha1.Instance)
	return instance.DeepCopyObject()
}

// Equivalent checks if two instance objects are the same
func (a *InstanceAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	instance1 := obj1.(*ocicorev1alpha1.Instance)
	instance2 := obj2.(*ocicorev1alpha1.Instance)
	if instance1.Status.Resource != nil {
		instance1.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	if instance2.Status.Resource != nil {
		instance2.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	if instance1.Status.PrimaryVnic != nil {
		instance1.Status.PrimaryVnic.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	if instance2.Status.PrimaryVnic != nil {
		instance2.Status.PrimaryVnic.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	if instance1.Status.BootVolume != nil {
		instance1.Status.BootVolume.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	if instance2.Status.BootVolume != nil {
		instance2.Status.BootVolume.TimeCreated = &ocisdkcommon.SDKTime{}
	}

	return reflect.DeepEqual(instance1, instance2)
}

// IsResourceCompliant checks if resource config is complient with CRD spec
func (a *InstanceAdapter) IsResourceCompliant(obj runtime.Object) bool {
	instance := obj.(*ocicorev1alpha1.Instance)

	if instance.Status.Resource == nil {
		return false
	}

	resource := instance.Status.Resource

	if resource.LifecycleState == ocicore.InstanceLifecycleStateStopped ||
		resource.LifecycleState == ocicore.InstanceLifecycleStateTerminating ||
		resource.LifecycleState == ocicore.InstanceLifecycleStateTerminated {
		return false
	}

	specDisplayName := resourcescommon.Display(instance.Name, instance.Spec.DisplayName)

	if *resource.DisplayName != *specDisplayName ||
		*resource.AvailabilityDomain != instance.Spec.AvailabilityDomain ||
		*resource.Shape != instance.Spec.Shape {
		return false
	}

	if instance.Spec.IpxeScript != "" && instance.Spec.IpxeScript != *resource.IpxeScript {
		return false
	}

	if instance.Spec.Metadata != nil && !reflect.DeepEqual(instance.Spec.Metadata, resource.Metadata) {
		return false
	}

	if instance.Spec.ExtendedMetadata != nil && !reflect.DeepEqual(instance.Spec.ExtendedMetadata, resource.ExtendedMetadata) {
		return false
	}

	return true
}

// IsResourceStatusChanged checks if two vcn objects are the same
func (a *InstanceAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	instance1 := obj1.(*ocicorev1alpha1.Instance)
	instance2 := obj2.(*ocicorev1alpha1.Instance)

	if (instance1.Status.BootVolume == nil && instance2.Status.BootVolume != nil) ||
		(instance2.Status.BootVolume == nil && instance1.Status.BootVolume != nil) {
		return true
	}

	if (instance1.Status.PrimaryVnic == nil && instance2.Status.PrimaryVnic != nil) ||
		(instance2.Status.PrimaryVnic == nil && instance1.Status.PrimaryVnic != nil) {
		return true
	}

	return instance1.Status.Resource.LifecycleState != instance2.Status.Resource.LifecycleState
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *InstanceAdapter) Id(obj runtime.Object) string {
	return obj.(*ocicorev1alpha1.Instance).GetResourceID()
}

// ObjectMeta returns the object meta struct from the instance object
func (a *InstanceAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ocicorev1alpha1.Instance).ObjectMeta
}

// DependsOn returns a map of instance dependencies (objects that the instance depends on)
func (a *InstanceAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ocicorev1alpha1.Instance).Spec.DependsOn
}

// Dependents returns a map of instance dependents (objects that depend on the instance)
func (a *InstanceAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ocicorev1alpha1.Instance).Status.Dependents
}

// CreateObject creates the instance object
func (a *InstanceAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Instance)
	return a.clientset.OcicoreV1alpha1().Instances(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the instance object
func (a *InstanceAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Instance)
	return a.clientset.OcicoreV1alpha1().Instances(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the instance object
func (a *InstanceAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var object = obj.(*ocicorev1alpha1.Instance)
	return a.clientset.OcicoreV1alpha1().Instances(object.ObjectMeta.Namespace).Delete(object.Name, options)
}

// DependsOnRefs returns the objects that the instance depends on
func (a *InstanceAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	var instance = obj.(*ocicorev1alpha1.Instance)

	deps := make([]runtime.Object, 0)

	if !resourcescommon.IsOcid(instance.Spec.CompartmentRef) {
		compartment, err := resourcescommon.Compartment(a.clientset, instance.ObjectMeta.Namespace, instance.Spec.CompartmentRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, compartment)
	}

	if !resourcescommon.IsOcid(instance.Spec.SubnetRef) {
		subnet, err := resourcescommon.Subnet(a.clientset, instance.ObjectMeta.Namespace, instance.Spec.SubnetRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, subnet)
	}
	return deps, nil
}

// PrimaryVnic returns the primary vnic resource of the instance
func (a *InstanceAdapter) PrimaryVnic(obj runtime.Object) (*ocicorev1alpha1.PrimaryVnicResource, error) {
	var object = obj.(*ocicorev1alpha1.Instance)

	request := ocicore.ListVnicAttachmentsRequest{}
	request.InstanceId = object.Status.Resource.Id
	request.CompartmentId = object.Status.Resource.CompartmentId

	r, err := a.cClient.ListVnicAttachments(a.ctx, request)

	if err != nil {
		return nil, err
	}

	for _, ociVnicAttachment := range r.Items {
		if ociVnicAttachment.LifecycleState == ocicore.VnicAttachmentLifecycleStateAttached {
			ociVnicResp, e := a.vcnClient.GetVnic(a.ctx, ocicore.GetVnicRequest{VnicId: ociVnicAttachment.VnicId})
			if e == nil && *ociVnicResp.Vnic.IsPrimary {
				return &ocicorev1alpha1.PrimaryVnicResource{Vnic: ociVnicResp.Vnic}, nil
			}
		}
	}

	return nil, errors.New("Primary Vnic not found")

}

// BootVolume returns the boot volume resource of the instance
func (a *InstanceAdapter) BootVolume(obj runtime.Object) (resource *ocicorev1alpha1.BootVolumeResource, err error) {
	var object = obj.(*ocicorev1alpha1.Instance)

	areq := ocicore.ListBootVolumeAttachmentsRequest{}
	areq.AvailabilityDomain = ocisdkcommon.String(object.Spec.AvailabilityDomain)
	areq.CompartmentId = object.Status.Resource.CompartmentId
	areq.InstanceId = object.Status.Resource.Id

	aresp, err := a.cClient.ListBootVolumeAttachments(a.ctx, areq)

	if err != nil {
		return nil, err
	}

	if len(aresp.Items) == 0 {
		return nil, errors.New("Can not find Boot Volume Attachment")
	}

	bootAttachment := aresp.Items[0]

	bvreq := ocicore.GetBootVolumeRequest{
		BootVolumeId: bootAttachment.BootVolumeId,
	}

	bvresp, err := a.bsClient.GetBootVolume(a.ctx, bvreq)

	if err != nil {
		return nil, err
	}
	return &ocicorev1alpha1.BootVolumeResource{BootVolume: bvresp.BootVolume}, nil
}

// Create creates the instance resource in oci
func (a *InstanceAdapter) Create(obj runtime.Object) (runtime.Object, error) {

	var (
		instance      = obj.(*ocicorev1alpha1.Instance)
		compartmentId *string
		imageId       *string
		subnetId      string
		err           error
	)

	if resourcescommon.IsOcid(instance.Spec.CompartmentRef) {

		compartmentId = ocisdkcommon.String(instance.Spec.CompartmentRef)
		imageId, err = a.getImageId(compartmentId, instance.Spec.Image)
		if err != nil {
			return instance, instance.Status.HandleError(err)
		}

	} else {

		compartment, err := resourcescommon.Compartment(a.clientset, instance.ObjectMeta.Namespace, instance.Spec.CompartmentRef)
		if err != nil {
			return instance, instance.Status.HandleError(err)
		}

		image := compartment.Status.Images[instance.Spec.Image]
		if image == "" {
			return instance, instance.Status.HandleError(errors.New("Unknown image specification"))
		}

		compartmentId = compartment.Status.Resource.Id
		imageId = ocisdkcommon.String(image)
	}

	if resourcescommon.IsOcid(instance.Spec.SubnetRef) {
		subnetId = instance.Spec.SubnetRef
	} else {
		subnetId, err = resourcescommon.SubnetId(a.clientset, instance.ObjectMeta.Namespace, instance.Spec.SubnetRef)
		if err != nil {
			return instance, instance.Status.HandleError(err)
		}
	}

	opcRetryToken := ocisdkcommon.String(string(instance.UID) + "-" + strconv.Itoa(instance.Status.ResetCounter))

	request := ocicore.LaunchInstanceRequest{}
	request.CompartmentId = compartmentId
	request.DisplayName = resourcescommon.Display(instance.Name, instance.Spec.DisplayName)
	request.AvailabilityDomain = ocisdkcommon.String(instance.Spec.AvailabilityDomain)
	request.ImageId = imageId
	request.Shape = ocisdkcommon.String(instance.Spec.Shape)
	request.SubnetId = ocisdkcommon.String(subnetId)
	request.HostnameLabel = resourcescommon.StrPtrOrNil(instance.Spec.HostnameLabel)
	request.IpxeScript = resourcescommon.StrPtrOrNil(instance.Spec.IpxeScript)
	request.Metadata = instance.Spec.Metadata
	request.ExtendedMetadata = instance.Spec.ExtendedMetadata
	request.OpcRetryToken = opcRetryToken

	r, err := a.cClient.LaunchInstance(a.ctx, request)

	if err != nil {
		return instance, instance.Status.HandleError(err)
	}

	return instance.SetResource(&r.Instance), instance.Status.HandleError(err)

}

func (a *InstanceAdapter) getImageId(compartmentId *string, imageName string) (*string, error) {
	request := ocicore.ListImagesRequest{
		CompartmentId: compartmentId,
	}

	r, err := a.cClient.ListImages(a.ctx, request)

	if r.Items == nil || len(r.Items) == 0 || err != nil {
		glog.Errorf("Invalid response from ListImages, error: %v", err)
		return nil, err
	}

	for _, ociImage := range r.Items {
		if *(ociImage.DisplayName) == imageName {
			return ociImage.Id, nil
		}
	}
	return nil, fmt.Errorf("Image not found")
}

// Delete deletes the instance resource in oci
func (a *InstanceAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Instance)

	rr := ocicore.GetInstanceRequest{
		InstanceId: object.Status.Resource.Id,
	}

	r, e := a.cClient.GetInstance(a.ctx, rr)

	if e != nil {
		return object, object.Status.HandleError(e)
	}

	if r.LifecycleState == ocicore.InstanceLifecycleStateTerminating {
		object.Status.State = ocicommon.ResourceStatePending
		return object, nil
	} else if r.LifecycleState == ocicore.InstanceLifecycleStateTerminated {
		object.Status.State = ocicommon.ResourceStateProcessed
		return object, nil
	}

	request := ocicore.TerminateInstanceRequest{
		InstanceId: object.Status.Resource.Id,
	}

	_, e = a.cClient.TerminateInstance(a.ctx, request)

	if e != nil {
		return object, object.Status.HandleError(e)
	}

	object.Status.State = ocicommon.ResourceStatePending
	return object, nil

}

// Get retrieves the instance resource from oci
func (a *InstanceAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Instance)

	request := ocicore.GetInstanceRequest{
		InstanceId: object.Status.Resource.Id,
	}

	r, e := a.cClient.GetInstance(a.ctx, request)

	if e != nil {
		return object, object.Status.HandleError(e)
	}

	object.SetResource(&r.Instance)

	if !(object.Status.PrimaryVnic != nil && object.Status.PrimaryVnic.LifecycleState == ocicore.VnicLifecycleStateAvailable) {
		vnic, verr := a.PrimaryVnic(object)
		if verr != nil {
			return object, object.Status.HandleError(verr)
		}
		object.Status.PrimaryVnic = vnic
	}

	if !(object.Status.BootVolume != nil && object.Status.BootVolume.LifecycleState == ocicore.BootVolumeLifecycleStateAvailable) {
		bootVol, bverr := a.BootVolume(object)
		if bverr != nil {
			return object, object.Status.HandleError(bverr)
		}
		object.Status.BootVolume = bootVol
	}

	return object, object.Status.HandleError(e)

}

// Update updates the instance resource in oci
func (a *InstanceAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Instance)

	if object.Status.Resource.LifecycleState == ocicore.InstanceLifecycleStateTerminated ||
		object.Status.Resource.LifecycleState == ocicore.InstanceLifecycleStateTerminating {
		glog.V(1).Infof("Got instance in %s state recreating: %s %#v\n", object.Status.Resource.LifecycleState, object.Name, object)
		object.Status.ResetCounter++
		object.Status.Resource = nil
		return object, nil

	} else if object.Status.Resource.LifecycleState == ocicore.InstanceLifecycleStateStopped {
		//bring it back up
		glog.V(1).Infof("Got instance in state STOPPED but needs to be RUNNING will bring it back up %s %#v\n", object.Name, object)
		r, e := a.cClient.InstanceAction(a.ctx, ocicore.InstanceActionRequest{InstanceId: object.Status.Resource.Id, Action: ocicore.InstanceActionActionStart})
		object.SetResource(&r.Instance)
		return object, object.Status.HandleError(e)

	}

	r, e := a.cClient.UpdateInstance(a.ctx, ocicore.UpdateInstanceRequest{InstanceId: object.Status.Resource.Id})

	if e != nil {
		return object, object.Status.HandleError(e)
	}

	object.SetResource(&r.Instance)
	return object, object.Status.HandleError(e)

}

// UpdateForResource calls a common UpdateForResource method to update the instance resource in the instance object
func (a *InstanceAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
