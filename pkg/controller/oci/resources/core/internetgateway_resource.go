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
		ocicorev1alpha1.InternetGatewayKind,
		ocicorev1alpha1.InternetGatewayResourcePlural,
		ocicorev1alpha1.InternetGatewayControllerName,
		&ocicorev1alpha1.InternetGatewayValidation,
		NewInternetGatewayAdapter)
}

// InternetGatewayAdapter implements the adapter interface for internet gateway resource
type InternetGatewayAdapter struct {
	clientset versioned.Interface
	ctx       context.Context
	vcnClient resourcescommon.VcnClientInterface
}

// NewInternetGatewayAdapter creates a new adapter for internet gateway resource
func NewInternetGatewayAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	iga := InternetGatewayAdapter{}

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
func (a *InternetGatewayAdapter) Kind() string {
	return ocicorev1alpha1.InternetGatewayKind
}

// Resource returns the plural name of the resource type
func (a *InternetGatewayAdapter) Resource() string {
	return ocicorev1alpha1.InternetGatewayResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *InternetGatewayAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ocicorev1alpha1.SchemeGroupVersion.WithResource(ocicorev1alpha1.InternetGatewayResourcePlural)
}

// ObjectType returns the internet gateway type for this adapter
func (a *InternetGatewayAdapter) ObjectType() runtime.Object {
	return &ocicorev1alpha1.InternetGateway{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *InternetGatewayAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ocicorev1alpha1.InternetGateway)
	return ok
}

// Copy returns a copy of a internet gateway object
func (a *InternetGatewayAdapter) Copy(obj runtime.Object) runtime.Object {
	internetgateway := obj.(*ocicorev1alpha1.InternetGateway)
	return internetgateway.DeepCopyObject()
}

// Equivalent checks if two internet gateway objects are the same
func (a *InternetGatewayAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	internetgateway1 := obj1.(*ocicorev1alpha1.InternetGateway)
	internetgateway2 := obj2.(*ocicorev1alpha1.InternetGateway)
	if internetgateway1.Status.Resource != nil {
		internetgateway1.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	if internetgateway2.Status.Resource != nil {
		internetgateway2.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	return reflect.DeepEqual(internetgateway1, internetgateway2)
}

// IsResourceCompliant checks if resource config is complient with CRD spec
func (a *InternetGatewayAdapter) IsResourceCompliant(obj runtime.Object) bool {
	ig := obj.(*ocicorev1alpha1.InternetGateway)

	if ig.Status.Resource == nil {
		return false
	}

	resource := ig.Status.Resource
	if resource.LifecycleState == ocicore.InternetGatewayLifecycleStateTerminating ||
		resource.LifecycleState == ocicore.InternetGatewayLifecycleStateProvisioning {
		return true
	}

	if resource.LifecycleState == ocicore.InternetGatewayLifecycleStateTerminated {
		return false
	}

	specDisplayName := resourcescommon.Display(ig.Name, ig.Spec.DisplayName)

	if *ig.Status.Resource.DisplayName != *specDisplayName ||
		*ig.Status.Resource.IsEnabled != ig.Spec.IsEnabled {
		return false
	}

	return true

}

// IsResourceStatusChanged checks if two vcn objects are the same
func (a *InternetGatewayAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	internetgateway1 := obj1.(*ocicorev1alpha1.InternetGateway)
	internetgateway2 := obj2.(*ocicorev1alpha1.InternetGateway)

	return internetgateway1.Status.Resource.LifecycleState != internetgateway2.Status.Resource.LifecycleState
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *InternetGatewayAdapter) Id(obj runtime.Object) string {
	return obj.(*ocicorev1alpha1.InternetGateway).GetResourceID()
}

// ObjectMeta returns the object meta struct from the internet gateway object
func (a *InternetGatewayAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ocicorev1alpha1.InternetGateway).ObjectMeta
}

// DependsOn returns a map of internet gateway dependencies (objects that the internet gateway depends on)
func (a *InternetGatewayAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ocicorev1alpha1.InternetGateway).Spec.DependsOn
}

// Dependents returns a map of internet gateway dependents (objects that depend on the internet gateway)
func (a *InternetGatewayAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ocicorev1alpha1.InternetGateway).Status.Dependents
}

// CreateObject creates the internet gateway object
func (a *InternetGatewayAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.InternetGateway)
	return a.clientset.OcicoreV1alpha1().InternetGatewaies(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the internet gateway object
func (a *InternetGatewayAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.InternetGateway)
	return a.clientset.OcicoreV1alpha1().InternetGatewaies(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the internet gateway object
func (a *InternetGatewayAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var object = obj.(*ocicorev1alpha1.InternetGateway)
	return a.clientset.OcicoreV1alpha1().InternetGatewaies(object.ObjectMeta.Namespace).Delete(object.Name, options)
}

// DependsOnRefs returns the objects that the internet gateway depends on
func (a *InternetGatewayAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	var ig = obj.(*ocicorev1alpha1.InternetGateway)
	deps := make([]runtime.Object, 0)

	if !resourcescommon.IsOcid(ig.Spec.CompartmentRef) {
		compartment, err := resourcescommon.Compartment(a.clientset, ig.ObjectMeta.Namespace, ig.Spec.CompartmentRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, compartment)
	}

	if !resourcescommon.IsOcid(ig.Spec.VcnRef) {
		virtualnetwork, err := resourcescommon.Vcn(a.clientset, ig.ObjectMeta.Namespace, ig.Spec.VcnRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, virtualnetwork)
	}
	return deps, nil
}

// Create creates the internet gateway resource in oci
func (a *InternetGatewayAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	var (
		ig               = obj.(*ocicorev1alpha1.InternetGateway)
		compartmentId    string
		virtualnetworkId string
		err              error
	)

	if resourcescommon.IsOcid(ig.Spec.CompartmentRef) {
		compartmentId = ig.Spec.CompartmentRef
	} else {
		compartmentId, err = resourcescommon.CompartmentId(a.clientset, ig.ObjectMeta.Namespace, ig.Spec.CompartmentRef)
		if err != nil {
			return ig, ig.Status.HandleError(err)
		}
	}

	if resourcescommon.IsOcid(ig.Spec.VcnRef) {
		virtualnetworkId = ig.Spec.VcnRef
	} else {
		virtualnetworkId, err = resourcescommon.VcnId(a.clientset, ig.ObjectMeta.Namespace, ig.Spec.VcnRef)
		if err != nil {
			return ig, ig.Status.HandleError(err)
		}
	}

	request := ocicore.CreateInternetGatewayRequest{}
	request.CompartmentId = ocisdkcommon.String(compartmentId)
	request.VcnId = ocisdkcommon.String(virtualnetworkId)
	request.DisplayName = resourcescommon.Display(ig.Name, ig.Spec.DisplayName)
	request.IsEnabled = ocisdkcommon.Bool(ig.Spec.IsEnabled)

	request.OpcRetryToken = ocisdkcommon.String(string(ig.UID))
	glog.Infof("InternetGateway: %s OpcRetryToken: %s", ig.Name, string(ig.UID))

	r, err := a.vcnClient.CreateInternetGateway(a.ctx, request)

	if err != nil {
		return ig, ig.Status.HandleError(err)
	}
	return ig.SetResource(&r.InternetGateway), ig.Status.HandleError(err)
}

// Delete deletes the internet gateway resource in oci
func (a *InternetGatewayAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.InternetGateway)

	request := ocicore.DeleteInternetGatewayRequest{
		IgId: object.Status.Resource.Id,
	}

	_, e := a.vcnClient.DeleteInternetGateway(a.ctx, request)

	if e == nil && object.Status.Resource != nil {
		object.Status.Resource.Id = ocisdkcommon.String("")
	}
	return object, object.Status.HandleError(e)
}

// Get retrieves the internet gateway resource from oci
func (a *InternetGatewayAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.InternetGateway)

	request := ocicore.GetInternetGatewayRequest{
		IgId: object.Status.Resource.Id,
	}

	e := wait.PollImmediate(1*time.Second, 10*time.Second, func() (bool, error) {
		r, e := a.vcnClient.GetInternetGateway(a.ctx, request)
		if e != nil {
			return false, e
		}
		if r.LifecycleState != ocicore.InternetGatewayLifecycleStateProvisioning {
			object.SetResource(&r.InternetGateway)
			return true, nil
		}
		return false, e
	})

	return object, object.Status.HandleError(e)
}

// Update updates the internet gateway resource in oci
func (a *InternetGatewayAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.InternetGateway)

	request := ocicore.UpdateInternetGatewayRequest{
		IgId: object.Status.Resource.Id,
	}

	if object.Status.Resource.LifecycleState != ocicore.InternetGatewayLifecycleStateAvailable {
		return object, errors.New(string(object.Status.Resource.LifecycleState))
	}

	r, e := a.vcnClient.UpdateInternetGateway(a.ctx, request)

	if e != nil {
		return object, object.Status.HandleError(e)
	}

	return object.SetResource(&r.InternetGateway), object.Status.HandleError(e)
}

// UpdateForResource calls a common UpdateForResource method to update the internet gateway resource in the internet gateway object
func (a *InternetGatewayAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
