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

	"github.com/golang/glog"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	resourcescommon.RegisterResourceTypeWithValidation(
		coregroup.GroupName,
		ocicorev1alpha1.RouteTableKind,
		ocicorev1alpha1.RouteTableResourcePlural,
		ocicorev1alpha1.RouteTableControllerName,
		&ocicorev1alpha1.RouteTableValidation,
		NewRouteTableAdapter)
}

// RouteTableAdapter implements the adapter interface for route table resource
type RouteTableAdapter struct {
	clientset versioned.Interface
	ctx       context.Context
	vcnClient resourcescommon.VcnClientInterface
}

// NewRouteTableAdapter creates a new adapter for route table resource
func NewRouteTableAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	rta := RouteTableAdapter{}
	vcnClient, err := ocicore.NewVirtualNetworkClientWithConfigurationProvider(ociconfig)

	if err != nil {
		glog.Errorf("Error creating oci VCN client: %v", err)
		os.Exit(1)
	}

	rta.vcnClient = &vcnClient
	rta.clientset = clientset
	rta.ctx = context.Background()

	return &rta
}

// Kind returns the resource kind string
func (a *RouteTableAdapter) Kind() string {
	return ocicorev1alpha1.RouteTableKind
}

// Resource returns the plural name of the resource type
func (a *RouteTableAdapter) Resource() string {
	return ocicorev1alpha1.RouteTableResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *RouteTableAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ocicorev1alpha1.SchemeGroupVersion.WithResource(ocicorev1alpha1.RouteTableResourcePlural)
}

// ObjectType returns the route table type for this adapter
func (a *RouteTableAdapter) ObjectType() runtime.Object {
	return &ocicorev1alpha1.RouteTable{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *RouteTableAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ocicorev1alpha1.RouteTable)
	return ok
}

// Copy returns a copy of a route table object
func (a *RouteTableAdapter) Copy(obj runtime.Object) runtime.Object {
	routetable := obj.(*ocicorev1alpha1.RouteTable)
	return routetable.DeepCopyObject()
}

// Equivalent checks if two route table objects are the same
func (a *RouteTableAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	routetable1 := obj1.(*ocicorev1alpha1.RouteTable)
	routetable2 := obj2.(*ocicorev1alpha1.RouteTable)
	if routetable1.Status.Resource != nil {
		routetable1.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	if routetable2.Status.Resource != nil {
		routetable2.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	return reflect.DeepEqual(routetable1, routetable2)
}

// IsResourceCompliant checks if resource config is complient with CRD spec
func (a *RouteTableAdapter) IsResourceCompliant(obj runtime.Object) bool {

	routetable := obj.(*ocicorev1alpha1.RouteTable)

	if routetable.Status.Resource == nil {
		return false
	}

	resource := routetable.Status.Resource

	if resource.LifecycleState == ocicore.RouteTableLifecycleStateProvisioning ||
		resource.LifecycleState == ocicore.RouteTableLifecycleStateTerminating {
		return true
	}

	if resource.LifecycleState == ocicore.RouteTableLifecycleStateTerminated {
		return false
	}

	specDisplayName := resourcescommon.Display(routetable.Name, routetable.Spec.DisplayName)

	if *routetable.Status.Resource.DisplayName != *specDisplayName {
		return false
	}

	specCidrBlocks := make(map[string]bool)
	resourceCidrBlocks := make(map[string]bool)

	for _, routeRule := range routetable.Spec.RouteRules {
		specCidrBlocks[routeRule.CidrBlock] = true
	}

	for _, routeRule := range routetable.Status.Resource.RouteRules {
		resourceCidrBlocks[*routeRule.CidrBlock] = true
	}

	return reflect.DeepEqual(specCidrBlocks, resourceCidrBlocks)

}

// IsResourceStatusChanged checks if two vcn objects are the same
func (a *RouteTableAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	routetable1 := obj1.(*ocicorev1alpha1.RouteTable)
	routetable2 := obj2.(*ocicorev1alpha1.RouteTable)

	return routetable1.Status.Resource.LifecycleState != routetable2.Status.Resource.LifecycleState
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *RouteTableAdapter) Id(obj runtime.Object) string {
	return obj.(*ocicorev1alpha1.RouteTable).GetResourceID()
}

// ObjectMeta returns the object meta struct from the route table object
func (a *RouteTableAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ocicorev1alpha1.RouteTable).ObjectMeta
}

// DependsOn returns a map of route table dependencies (objects that the route table depends on)
func (a *RouteTableAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ocicorev1alpha1.RouteTable).Spec.DependsOn
}

// Dependents returns a map of route table dependents (objects that depend on the route table)
func (a *RouteTableAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ocicorev1alpha1.RouteTable).Status.Dependents
}

// CreateObject creates the route table object
func (a *RouteTableAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.RouteTable)
	return a.clientset.OcicoreV1alpha1().RouteTables(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the route table object
func (a *RouteTableAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.RouteTable)
	return a.clientset.OcicoreV1alpha1().RouteTables(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the route table object
func (a *RouteTableAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var object = obj.(*ocicorev1alpha1.RouteTable)
	return a.clientset.OcicoreV1alpha1().RouteTables(object.ObjectMeta.Namespace).Delete(object.Name, options)
}

// DependsOnRefs returns the objects that the route table depends on
func (a *RouteTableAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.RouteTable)
	deps := make([]runtime.Object, 0)

	if !resourcescommon.IsOcid(object.Spec.CompartmentRef) {
		compartment, err := resourcescommon.Compartment(a.clientset, object.ObjectMeta.Namespace, object.Spec.CompartmentRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, compartment)
	}

	if !resourcescommon.IsOcid(object.Spec.VcnRef) {
		virtualnetwork, err := resourcescommon.Vcn(a.clientset, object.ObjectMeta.Namespace, object.Spec.VcnRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, virtualnetwork)
	}

	for _, routeRule := range object.Spec.RouteRules {
		if !resourcescommon.IsOcid(routeRule.NetworkEntityID) {
			internetgateway, err := resourcescommon.InternetGateway(a.clientset, object.ObjectMeta.Namespace, routeRule.NetworkEntityID)
			if err != nil {
				return nil, err
			}
			deps = append(deps, internetgateway)
		}
	}
	return deps, nil
}

// Create creates the route table resource in oci
func (a *RouteTableAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	var (
		object           = obj.(*ocicorev1alpha1.RouteTable)
		compartmentId    string
		virtualnetworkId string
		err              error
	)

	if resourcescommon.IsOcid(object.Spec.CompartmentRef) {
		compartmentId = object.Spec.CompartmentRef
	} else {
		compartmentId, err = resourcescommon.CompartmentId(a.clientset, object.ObjectMeta.Namespace, object.Spec.CompartmentRef)
		if err != nil {
			return object, object.Status.HandleError(err)
		}
	}

	if resourcescommon.IsOcid(object.Spec.VcnRef) {
		virtualnetworkId = object.Spec.VcnRef
	} else {
		virtualnetworkId, err = resourcescommon.VcnId(a.clientset, object.ObjectMeta.Namespace, object.Spec.VcnRef)
		if err != nil {
			return object, object.Status.HandleError(err)
		}
	}

	var routeRuleList []ocicore.RouteRule
	for _, routeRule := range object.Spec.RouteRules {
		// TODO add other network entity types
		var internetgatewayId string
		if resourcescommon.IsOcid(routeRule.NetworkEntityID) {
			internetgatewayId = routeRule.NetworkEntityID
		} else {
			internetgatewayId, err = resourcescommon.InternetGatewayId(a.clientset, object.ObjectMeta.Namespace, routeRule.NetworkEntityID)
			if err != nil {
				return object, object.Status.HandleError(err)
			}
		}
		routeRuleList = append(routeRuleList, ocicore.RouteRule{
			CidrBlock:       ocisdkcommon.String(routeRule.CidrBlock),
			NetworkEntityId: ocisdkcommon.String(internetgatewayId),
		},
		)
	}

	// create a new RouteTable
	request := ocicore.CreateRouteTableRequest{}
	request.CompartmentId = ocisdkcommon.String(compartmentId)
	request.VcnId = ocisdkcommon.String(virtualnetworkId)
	request.RouteRules = routeRuleList
	request.DisplayName = resourcescommon.Display(object.Name, object.Spec.DisplayName)

	request.OpcRetryToken = ocisdkcommon.String(string(object.UID))
	glog.Infof("RouteTable: %s OpcRetryToken: %s", object.Name, string(object.UID))

	r, err := a.vcnClient.CreateRouteTable(a.ctx, request)

	if err != nil {
		return object, object.Status.HandleError(err)
	}

	return object.SetResource(&r.RouteTable), object.Status.HandleError(err)
}

// Delete deletes the route table resource in oci
func (a *RouteTableAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.RouteTable)
	request := ocicore.DeleteRouteTableRequest{
		RtId: object.Status.Resource.Id,
	}

	_, e := a.vcnClient.DeleteRouteTable(a.ctx, request)

	if e == nil && object.Status.Resource != nil {
		object.Status.Resource.Id = ocisdkcommon.String("")
	}
	return object, object.Status.HandleError(e)
}

// Get retrieves the route table resource from oci
func (a *RouteTableAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.RouteTable)

	request := ocicore.GetRouteTableRequest{
		RtId: object.Status.Resource.Id,
	}

	r, e := a.vcnClient.GetRouteTable(a.ctx, request)

	if e != nil {
		return object, object.Status.HandleError(e)
	}

	return object.SetResource(&r.RouteTable), object.Status.HandleError(e)
}

// Update updates the route table resource in oci
func (a *RouteTableAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.RouteTable)

	request := ocicore.UpdateRouteTableRequest{
		RtId: object.Status.Resource.Id,
	}

	r, e := a.vcnClient.UpdateRouteTable(a.ctx, request)

	if e != nil {
		return object, object.Status.HandleError(e)
	}

	return object.SetResource(&r.RouteTable), object.Status.HandleError(e)
}

// UpdateForResource calls a common UpdateForResource method to update the route table resource in the route table object
func (a *RouteTableAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
