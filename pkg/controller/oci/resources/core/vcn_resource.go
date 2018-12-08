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
	"k8s.io/client-go/kubernetes"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ocicommon "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	coregroup "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com"
	ocicorev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"

	"context"
	"os"

	"github.com/golang/glog"
	ocisdkcommon "github.com/oracle/oci-go-sdk/common"
	ocicore "github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	resourcescommon.RegisterResourceTypeWithValidation(
		coregroup.GroupName,
		ocicorev1alpha1.VirtualNetworkKind,
		ocicorev1alpha1.VirtualNetworkResourcePlural,
		ocicorev1alpha1.VirtualNetworkControllerName,
		&ocicorev1alpha1.VcnValidation,
		NewVcnAdapter)
}

// VcnAdapter implements the adapter interface for vcn resource
type VcnAdapter struct {
	clientset versioned.Interface
	vcnClient resourcescommon.VcnClientInterface
	ctx       context.Context
}

// NewVcnAdapter creates a new adapter for vcn resource
func NewVcnAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	vna := VcnAdapter{}

	vcnClient, err := ocicore.NewVirtualNetworkClientWithConfigurationProvider(ociconfig)

	if err != nil {
		glog.Errorf("Error creating oci VCN client: %v", err)
		os.Exit(1)
	}

	vna.vcnClient = &vcnClient
	vna.clientset = clientset
	vna.ctx = context.Background()
	return &vna
}

func NewVcnAdapterBasic(clientset versioned.Interface, vcnClient resourcescommon.VcnClientInterface) resourcescommon.ResourceTypeAdapter {
	vna := VcnAdapter{}
	vna.vcnClient = vcnClient
	vna.clientset = clientset
	vna.ctx = context.Background()
	return &vna
}

// Kind returns the resource kind string
func (a *VcnAdapter) Kind() string {
	return ocicorev1alpha1.VirtualNetworkKind
}

// Resource returns the plural name of the resource type
func (a *VcnAdapter) Resource() string {
	return ocicorev1alpha1.VirtualNetworkResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *VcnAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ocicorev1alpha1.SchemeGroupVersion.WithResource(ocicorev1alpha1.VirtualNetworkResourcePlural)
}

// ObjectType returns the vcn type for this adapter
func (a *VcnAdapter) ObjectType() runtime.Object {
	return &ocicorev1alpha1.Vcn{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *VcnAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ocicorev1alpha1.Vcn)
	return ok
}

// Copy returns a copy of a vcn object
func (a *VcnAdapter) Copy(obj runtime.Object) runtime.Object {
	virtualnetwork := obj.(*ocicorev1alpha1.Vcn)
	return virtualnetwork.DeepCopyObject()
}

// Equivalent checks if two vcn objects are the same
func (a *VcnAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	virtualnetwork1 := obj1.(*ocicorev1alpha1.Vcn)
	virtualnetwork2 := obj2.(*ocicorev1alpha1.Vcn)
	if virtualnetwork1.Status.Resource != nil {
		virtualnetwork1.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	if virtualnetwork2.Status.Resource != nil {
		virtualnetwork2.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	return reflect.DeepEqual(virtualnetwork1, virtualnetwork2)
}

// IsResourceCompliant
func (a *VcnAdapter) IsResourceCompliant(obj runtime.Object) bool {
	virtualnetwork := obj.(*ocicorev1alpha1.Vcn)
	if virtualnetwork.Status.Resource == nil {
		return false
	}

	resource := virtualnetwork.Status.Resource

	if resource.LifecycleState == ocicore.VcnLifecycleStateProvisioning ||
		resource.LifecycleState == ocicore.VcnLifecycleStateTerminating {
		return true
	}

	if resource.LifecycleState == ocicore.VcnLifecycleStateTerminated {
		return false
	}

	displayName := resourcescommon.Display(virtualnetwork.Name, virtualnetwork.Spec.DisplayName)

	if *resource.CidrBlock != virtualnetwork.Spec.CidrBlock ||
		*resource.DisplayName != *displayName ||
		*resource.DnsLabel != virtualnetwork.Spec.DNSLabel {
		return false
	}
	return true
}

// IsResourceStatusChanged checks if two vcn objects are the same
func (a *VcnAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	virtualnetwork1 := obj1.(*ocicorev1alpha1.Vcn)
	virtualnetwork2 := obj2.(*ocicorev1alpha1.Vcn)

	if virtualnetwork1.Status.Resource.LifecycleState != virtualnetwork2.Status.Resource.LifecycleState {
		return true
	}

	return false
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *VcnAdapter) Id(obj runtime.Object) string {
	return obj.(*ocicorev1alpha1.Vcn).GetResourceID()
}

// ObjectMeta returns the object meta struct from the vcn object
func (a *VcnAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ocicorev1alpha1.Vcn).ObjectMeta
}

// DependsOn returns a map of vcn dependencies (objects that the vcn depends on)
func (a *VcnAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ocicorev1alpha1.Vcn).Spec.DependsOn
}

// Dependents returns a map of vcn dependents (objects that depend on the vcn)
func (a *VcnAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ocicorev1alpha1.Vcn).Status.Dependents
}

// CreateObject creates the vcn object
func (a *VcnAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Vcn)
	return a.clientset.OcicoreV1alpha1().Vcns(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the vcn object
func (a *VcnAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.Vcn)
	return a.clientset.OcicoreV1alpha1().Vcns(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the vcn object
func (a *VcnAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var object = obj.(*ocicorev1alpha1.Vcn)
	return a.clientset.OcicoreV1alpha1().Vcns(object.ObjectMeta.Namespace).Delete(object.Name, options)
}

// DependsOnRefs returns the objects that the vcn depends on
func (a *VcnAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	object := obj.(*ocicorev1alpha1.Vcn)
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

// Create creates the vcn resource in oci
func (a *VcnAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	object := obj.(*ocicorev1alpha1.Vcn)
	var compartmentId string

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
	// create a new VCN
	request := ocicore.CreateVcnRequest{}
	request.CidrBlock = ocisdkcommon.String(object.Spec.CidrBlock)
	request.CompartmentId = ocisdkcommon.String(compartmentId)
	request.DisplayName = resourcescommon.Display(object.Name, object.Spec.DisplayName)
	request.DnsLabel = ocisdkcommon.String(object.Spec.DNSLabel)

	request.OpcRetryToken = ocisdkcommon.String(string(object.UID))

	r, err := a.vcnClient.CreateVcn(a.ctx, request)

	if err != nil {
		return object, object.Status.HandleError(err)
	}

	return object.SetResource(&r.Vcn), object.Status.HandleError(err)
}

// Delete deletes the vcn resource in oci
func (a *VcnAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	object := obj.(*ocicorev1alpha1.Vcn)
	request := ocicore.DeleteVcnRequest{
		VcnId: object.Status.Resource.Id,
	}

	_, e := a.vcnClient.DeleteVcn(a.ctx, request)

	if e == nil && object.Status.Resource != nil {
		object.Status.Resource.Id = ocisdkcommon.String("")
	}
	return object, object.Status.HandleError(e)
}

// Get retrieves the vcn resource from oci
func (a *VcnAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	object := obj.(*ocicorev1alpha1.Vcn)

	request := ocicore.GetVcnRequest{
		VcnId: object.Status.Resource.Id,
	}

	r, e := a.vcnClient.GetVcn(a.ctx, request)

	if e != nil {
		return object, object.Status.HandleError(e)
	}

	return object.SetResource(&r.Vcn), object.Status.HandleError(e)
}

// Update updates the vcn resource in oci
func (a *VcnAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	object := obj.(*ocicorev1alpha1.Vcn)

	request := ocicore.UpdateVcnRequest{
		VcnId: object.Status.Resource.Id,
	}

	r, e := a.vcnClient.UpdateVcn(a.ctx, request)

	if e != nil {
		return object, object.Status.HandleError(e)
	}

	return object.SetResource(&r.Vcn), object.Status.HandleError(e)
}

// UpdateForResource calls a common UpdateForResource method to update the vcn resource in the vcn object
func (a *VcnAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
