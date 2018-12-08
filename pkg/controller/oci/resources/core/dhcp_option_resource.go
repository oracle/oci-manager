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

	ocisdkcommon "github.com/oracle/oci-go-sdk/common"
	ocicore "github.com/oracle/oci-go-sdk/core"

	"context"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

func init() {
	resourcescommon.RegisterResourceTypeWithValidation(
		coregroup.GroupName,
		ocicorev1alpha1.DhcpOptionKind,
		ocicorev1alpha1.DhcpOptionResourcePlural,
		ocicorev1alpha1.DhcpOptionControllerName,
		&ocicorev1alpha1.DhcpOptionValidation,
		NewDhcpOptionAdapter)
}

// DhcpOptionAdapter implements the adapter interface for dhcp options resource
type DhcpOptionAdapter struct {
	clientset versioned.Interface
	ctx       context.Context
	vcnClient resourcescommon.VcnClientInterface
}

// NewDhcpOptionAdapter creates a new adapter for dhcp options resource
func NewDhcpOptionAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	iga := DhcpOptionAdapter{}

	vcnClient, err := ocicore.NewVirtualNetworkClientWithConfigurationProvider(ociconfig)

	if err != nil {
		glog.Errorf("Error creating oci VCN client: %v", err)
		os.Exit(1)
	}

	iga.vcnClient = &vcnClient
	iga.clientset = clientset
	iga.ctx = context.Background()

	return &iga
}

// Kind returns the resource kind string
func (a *DhcpOptionAdapter) Kind() string {
	return ocicorev1alpha1.DhcpOptionKind
}

// Resource returns the plural name of the resource type
func (a *DhcpOptionAdapter) Resource() string {
	return ocicorev1alpha1.DhcpOptionResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *DhcpOptionAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ocicorev1alpha1.SchemeGroupVersion.WithResource(ocicorev1alpha1.DhcpOptionResourcePlural)
}

// ObjectType returns the dhcp options type for this adapter
func (a *DhcpOptionAdapter) ObjectType() runtime.Object {
	return &ocicorev1alpha1.DhcpOption{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *DhcpOptionAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ocicorev1alpha1.DhcpOption)
	return ok
}

// Copy returns a copy of a dhcp options object
func (a *DhcpOptionAdapter) Copy(obj runtime.Object) runtime.Object {
	internetgateway := obj.(*ocicorev1alpha1.DhcpOption)
	return internetgateway.DeepCopyObject()
}

// Equivalent checks if two dhcp options objects are the same
func (a *DhcpOptionAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	dhcpOption1 := obj1.(*ocicorev1alpha1.DhcpOption)
	dhcpOption2 := obj2.(*ocicorev1alpha1.DhcpOption)
	if dhcpOption1.Status.Resource != nil {
		dhcpOption1.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	if dhcpOption2.Status.Resource != nil {
		dhcpOption2.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	return reflect.DeepEqual(dhcpOption1, dhcpOption2)
}

// IsResourceCompliant checks if resource config is complient with CRD spec
func (a *DhcpOptionAdapter) IsResourceCompliant(obj runtime.Object) bool {
	do := obj.(*ocicorev1alpha1.DhcpOption)

	if do.Status.Resource == nil {
		return false
	}

	resource := do.Status.Resource

	if resource.LifecycleState == ocicore.DhcpOptionsLifecycleStateProvisioning ||
		resource.LifecycleState == ocicore.DhcpOptionsLifecycleStateTerminating {
		return true
	}

	if resource.LifecycleState == ocicore.DhcpOptionsLifecycleStateTerminated {
		return false
	}

	specName := resourcescommon.Display(do.Name, do.Spec.DisplayName)

	if do.Status.Resource.DisplayName != specName {
		return false
	}

	return reflect.DeepEqual(do.Spec.Options, do.Status.Resource.Options)

}

// IsResourceStatusChanged checks if two objects are the same
func (a *DhcpOptionAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	do1 := obj1.(*ocicorev1alpha1.DhcpOption)
	do2 := obj2.(*ocicorev1alpha1.DhcpOption)

	return do1.Status.Resource.LifecycleState != do2.Status.Resource.LifecycleState
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *DhcpOptionAdapter) Id(obj runtime.Object) string {
	return obj.(*ocicorev1alpha1.DhcpOption).GetResourceID()
}

// ObjectMeta returns the object meta struct from the dhcp options object
func (a *DhcpOptionAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ocicorev1alpha1.DhcpOption).ObjectMeta
}

// DependsOn returns a map of dhcp options dependencies (objects that the dhcp options depends on)
func (a *DhcpOptionAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ocicorev1alpha1.DhcpOption).Spec.DependsOn
}

// Dependents returns a map of dhcp options dependents (objects that depend on the dhcp options)
func (a *DhcpOptionAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ocicorev1alpha1.DhcpOption).Status.Dependents
}

// CreateObject creates the dhcp options object
func (a *DhcpOptionAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.DhcpOption)
	return a.clientset.OcicoreV1alpha1().DhcpOptions(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the dhcp options object
func (a *DhcpOptionAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.DhcpOption)
	return a.clientset.OcicoreV1alpha1().DhcpOptions(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the dhcp options object
func (a *DhcpOptionAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var object = obj.(*ocicorev1alpha1.DhcpOption)
	return a.clientset.OcicoreV1alpha1().DhcpOptions(object.ObjectMeta.Namespace).Delete(object.Name, options)
}

// DependsOnRefs returns the objects that the dhcp options depends on
func (a *DhcpOptionAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	var do = obj.(*ocicorev1alpha1.DhcpOption)
	deps := make([]runtime.Object, 0)

	if !resourcescommon.IsOcid(do.Spec.CompartmentRef) {
		compartment, err := resourcescommon.Compartment(a.clientset, do.ObjectMeta.Namespace, do.Spec.CompartmentRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, compartment)
	}

	if !resourcescommon.IsOcid(do.Spec.VcnRef) {
		virtualnetwork, err := resourcescommon.Vcn(a.clientset, do.ObjectMeta.Namespace, do.Spec.VcnRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, virtualnetwork)
	}
	return deps, nil
}

// Create creates the dhcp options resource in oci
func (a *DhcpOptionAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	glog.Infof("in create")
	var (
		do            = obj.(*ocicorev1alpha1.DhcpOption)
		compartmentId string
		vcnId         string
		err           error
	)

	if resourcescommon.IsOcid(do.Spec.CompartmentRef) {
		compartmentId = do.Spec.CompartmentRef
	} else {
		compartmentId, err = resourcescommon.CompartmentId(a.clientset, do.ObjectMeta.Namespace, do.Spec.CompartmentRef)
		if err != nil {
			return do, do.Status.HandleError(err)
		}
	}

	if resourcescommon.IsOcid(do.Spec.VcnRef) {
		vcnId = do.Spec.VcnRef
	} else {
		vcnId, err = resourcescommon.VcnId(a.clientset, do.ObjectMeta.Namespace, do.Spec.VcnRef)
		if err != nil {
			return do, do.Status.HandleError(err)
		}
	}

	request := ocicore.CreateDhcpOptionsRequest{}
	request.CompartmentId = ocisdkcommon.String(compartmentId)
	request.VcnId = ocisdkcommon.String(vcnId)
	request.DisplayName = resourcescommon.Display(do.Name, do.Spec.DisplayName)
	request.Options = do.Spec.Options

	request.OpcRetryToken = ocisdkcommon.String(string(do.UID))

	r, err := a.vcnClient.CreateDhcpOptions(a.ctx, request)

	if err != nil {
		return do, do.Status.HandleError(err)
	}

	return do.SetResource(&r.DhcpOptions), do.Status.HandleError(err)
}

// Delete deletes the dhcp options resource in oci
func (a *DhcpOptionAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.DhcpOption)

	request := ocicore.DeleteDhcpOptionsRequest{
		DhcpId: object.Status.Resource.Id,
	}

	_, e := a.vcnClient.DeleteDhcpOptions(a.ctx, request)

	if e == nil && object.Status.Resource != nil {
		object.Status.Resource.Id = ocisdkcommon.String("")
	}
	return object, object.Status.HandleError(e)
}

// Get retrieves the dhcp options resource from oci
func (a *DhcpOptionAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.DhcpOption)

	request := ocicore.GetDhcpOptionsRequest{
		DhcpId: object.Status.Resource.Id,
	}

	e := wait.PollImmediate(1*time.Second, 10*time.Second, func() (bool, error) {
		r, e := a.vcnClient.GetDhcpOptions(a.ctx, request)
		if e != nil {
			return false, e
		}
		if r.LifecycleState != ocicore.DhcpOptionsLifecycleStateProvisioning {
			object.SetResource(&r.DhcpOptions)
			return true, nil
		}
		return false, e
	})

	if e != nil {
		return object, object.Status.HandleError(e)
	}

	return object, object.Status.HandleError(e)
}

// Update updates the dhcp options resource in oci
func (a *DhcpOptionAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.DhcpOption)

	details := ocicore.UpdateDhcpDetails{
		DisplayName: resourcescommon.Display(object.Name, object.Spec.DisplayName),
		Options:     object.Spec.Options,
	}

	request := ocicore.UpdateDhcpOptionsRequest{
		DhcpId:            object.Status.Resource.Id,
		UpdateDhcpDetails: details,
	}

	if object.Status.Resource.LifecycleState != ocicore.DhcpOptionsLifecycleStateAvailable {
		return object, errors.New(string(object.Status.Resource.LifecycleState))
	}

	r, e := a.vcnClient.UpdateDhcpOptions(a.ctx, request)

	if e != nil {
		return object, object.Status.HandleError(e)
	}

	return object.SetResource(&r.DhcpOptions), object.Status.HandleError(e)
}

// UpdateForResource calls a common UpdateForResource method to update the dhcp options resource in the dhcp options object
func (a *DhcpOptionAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
