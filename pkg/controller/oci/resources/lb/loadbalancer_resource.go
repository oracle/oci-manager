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

package lb

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"os"
	"reflect"

	ocisdkcommon "github.com/oracle/oci-go-sdk/common"
	ocisdklb "github.com/oracle/oci-go-sdk/loadbalancer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	ocicommon "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	lbgroup "github.com/oracle/oci-manager/pkg/apis/ocilb.oracle.com"
	ocilbv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocilb.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
)

func init() {
	resourcescommon.RegisterResourceTypeWithValidation(
		lbgroup.GroupName,
		ocilbv1alpha1.LoadBalancerKind,
		ocilbv1alpha1.LoadBalancerResourcePlural,
		ocilbv1alpha1.LoadBalancerControllerName,
		&ocilbv1alpha1.LoadBalancerValidation,
		NewLoadBalancerAdapter)
}

// LoadBalancerAdapter implements the adapter interface for load balancer resource
type LoadBalancerAdapter struct {
	clientset versioned.Interface
	ctx       context.Context
	lbClient  resourcescommon.LoadBalancerClientInterface
}

// NewLoadBalancerAdapter creates a new adapter for load balancer resource
func NewLoadBalancerAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	lba := LoadBalancerAdapter{}
	lba.clientset = clientset
	lba.ctx = context.Background()

	lbClient, err := ocisdklb.NewLoadBalancerClientWithConfigurationProvider(ociconfig)
	if err != nil {
		glog.Errorf("Error creating oci lb client: %v", err)
		os.Exit(1)
	}
	lba.lbClient = &lbClient
	return &lba
}

// Kind returns the resource kind string
func (a *LoadBalancerAdapter) Kind() string {
	return ocilbv1alpha1.LoadBalancerKind
}

// Resource returns the plural name of the resource type
func (a *LoadBalancerAdapter) Resource() string {
	return ocilbv1alpha1.LoadBalancerResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *LoadBalancerAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ocilbv1alpha1.SchemeGroupVersion.WithResource(ocilbv1alpha1.LoadBalancerResourcePlural)
}

// ObjectType returns the load balancer type for this adapter
func (a *LoadBalancerAdapter) ObjectType() runtime.Object {
	return &ocilbv1alpha1.LoadBalancer{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *LoadBalancerAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ocilbv1alpha1.LoadBalancer)
	return ok
}

// Copy returns a copy of a load balancer object
func (a *LoadBalancerAdapter) Copy(obj runtime.Object) runtime.Object {
	LoadBalancer := obj.(*ocilbv1alpha1.LoadBalancer)
	return LoadBalancer.DeepCopyObject()
}

// Equivalent checks if two load balancer objects are the same
func (a *LoadBalancerAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	LoadBalancer1 := obj1.(*ocilbv1alpha1.LoadBalancer)
	LoadBalancer2 := obj2.(*ocilbv1alpha1.LoadBalancer)
	if LoadBalancer1.Status.Resource != nil {
		LoadBalancer1.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	if LoadBalancer2.Status.Resource != nil {
		LoadBalancer2.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	return reflect.DeepEqual(LoadBalancer1, LoadBalancer2)
}

// IsResourceCompliant checks if resource config is complient with CRD spec
func (a *LoadBalancerAdapter) IsResourceCompliant(obj runtime.Object) bool {
	lb := obj.(*ocilbv1alpha1.LoadBalancer)
	if lb.Status.Resource == nil {
		return false
	}

	resource := lb.Status.Resource

	if resource.LifecycleState == ocisdklb.LoadBalancerLifecycleStateCreating ||
		resource.LifecycleState == ocisdklb.LoadBalancerLifecycleStateDeleting {
		return true
	}

	if resource.LifecycleState == ocisdklb.LoadBalancerLifecycleStateDeleted ||
		resource.LifecycleState == ocisdklb.LoadBalancerLifecycleStateFailed {
		return false
	}

	if lb.Spec.Shape != *resource.ShapeName ||
		lb.Spec.IsPrivate != *resource.IsPrivate {
		return false
	}
	return true
}

// IsResourceStatusChanged checks if two vcn objects are the same
func (a *LoadBalancerAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	lb1 := obj1.(*ocilbv1alpha1.LoadBalancer)
	lb2 := obj2.(*ocilbv1alpha1.LoadBalancer)

	return lb1.Status.Resource.LifecycleState != lb2.Status.Resource.LifecycleState
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *LoadBalancerAdapter) Id(obj runtime.Object) string {
	return obj.(*ocilbv1alpha1.LoadBalancer).GetResourceID()
}

// ObjectMeta returns the object meta struct from the load balancer object
func (a *LoadBalancerAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ocilbv1alpha1.LoadBalancer).ObjectMeta
}

// DependsOn returns a map of load balancer dependencies (objects that the load balancer depends on)
func (a *LoadBalancerAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ocilbv1alpha1.LoadBalancer).Spec.DependsOn
}

// Dependents returns a map of load balancer dependents (objects that depend on the load balancer)
func (a *LoadBalancerAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ocilbv1alpha1.LoadBalancer).Status.Dependents
}

// CreateObject creates the load balancer object
func (a *LoadBalancerAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocilbv1alpha1.LoadBalancer)
	return a.clientset.OcilbV1alpha1().LoadBalancers(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the load balancer object
func (a *LoadBalancerAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocilbv1alpha1.LoadBalancer)
	return a.clientset.OcilbV1alpha1().LoadBalancers(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the load balancer object
func (a *LoadBalancerAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var object = obj.(*ocilbv1alpha1.LoadBalancer)
	return a.clientset.OcilbV1alpha1().LoadBalancers(object.ObjectMeta.Namespace).Delete(object.Name, options)
}

// DependsOnRefs returns the objects that the load balancer depends on
func (a *LoadBalancerAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	var lb = obj.(*ocilbv1alpha1.LoadBalancer)

	deps := make([]runtime.Object, 0)

	if !resourcescommon.IsOcid(lb.Spec.CompartmentRef) {
		compartment, err := resourcescommon.Compartment(a.clientset, lb.ObjectMeta.Namespace, lb.Spec.CompartmentRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, compartment)
	}

	for _, subnetName := range lb.Spec.SubnetRefs {
		if !resourcescommon.IsOcid(subnetName) {
			subnet, err := resourcescommon.Subnet(a.clientset, lb.ObjectMeta.Namespace, subnetName)

			if err != nil {
				return nil, err
			}
			deps = append(deps, subnet)
		}
	}

	return deps, nil
}

// Create creates the load balancer resource in oci
func (a *LoadBalancerAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	var (
		lb            = obj.(*ocilbv1alpha1.LoadBalancer).DeepCopy()
		compartmentId string
		err           error
	)

	if lb.Status.WorkRequestId != nil {

		workRequest := ocisdklb.GetWorkRequestRequest{WorkRequestId: lb.Status.WorkRequestId}
		workResp, e := a.lbClient.GetWorkRequest(a.ctx, workRequest)
		if e != nil {
			glog.Errorf("CreateLoadBalancer GetWorkRequest error: %v", e)
			return lb, lb.Status.HandleError(e)
		}
		glog.Infof("CreateLoadBalancer workResp state: %s", workResp.LifecycleState)

		if workResp.LifecycleState != ocisdklb.WorkRequestLifecycleStateSucceeded &&
			workResp.LifecycleState != ocisdklb.WorkRequestLifecycleStateFailed {

			if lb.Status.WorkRequestStatus == nil ||
				workResp.LifecycleState != *lb.Status.WorkRequestStatus {
				lb.Status.WorkRequestStatus = &workResp.LifecycleState
				return lb, nil
			} else {
				return nil, nil
			}
		}

		if workResp.LifecycleState == ocisdklb.WorkRequestLifecycleStateFailed {
			lb.Status.WorkRequestStatus = &workResp.LifecycleState
			err := fmt.Errorf("WorkRequest %s is in failed state", *lb.Status.WorkRequestId)
			return lb, lb.Status.HandleError(err)
		}

		lb.Status.Resource = &ocilbv1alpha1.LoadBalancerResource{
			LoadBalancer: &ocisdklb.LoadBalancer{
				Id: workResp.LoadBalancerId,
			},
		}
		lb.Status.WorkRequestId = nil
		lb.Status.WorkRequestStatus = nil

	} else {
		if resourcescommon.IsOcid(lb.Spec.CompartmentRef) {
			compartmentId = lb.Spec.CompartmentRef
		} else {
			compartmentId, err = resourcescommon.CompartmentId(a.clientset, lb.ObjectMeta.Namespace, lb.Spec.CompartmentRef)
			if err != nil {
				return lb, lb.Status.HandleError(err)
			}
			glog.Infof("CreateLoadBalancer compartment id: %s", compartmentId)
		}

		subnets := make([]string, 0)
		for _, subnetName := range lb.Spec.SubnetRefs {
			if resourcescommon.IsOcid(subnetName) {
				subnets = append(subnets, subnetName)
			} else {
				subnetId, err := resourcescommon.SubnetId(a.clientset, lb.ObjectMeta.Namespace, subnetName)
				if err != nil {
					return lb, lb.Status.HandleError(err)
				}
				subnets = append(subnets, subnetId)
			}
		}
		glog.Infof("CreateLoadBalancer subnets: %s", subnets)

		createDetails := ocisdklb.CreateLoadBalancerDetails{
			CompartmentId: ocisdkcommon.String(compartmentId),
			DisplayName:   &lb.Name,
			IsPrivate:     &lb.Spec.IsPrivate,
			ShapeName:     &lb.Spec.Shape,
			SubnetIds:     subnets,
		}
		createRequest := ocisdklb.CreateLoadBalancerRequest{
			CreateLoadBalancerDetails: createDetails,
			OpcRetryToken:             ocisdkcommon.String(string(lb.UID)),
		}

		createResponse, e := a.lbClient.CreateLoadBalancer(a.ctx, createRequest)
		if e != nil {
			glog.Errorf("CreateLoadBalancer error: %v", e)
			return lb, lb.Status.HandleError(e)
		}
		glog.Infof("CreateLoadBalancer workRequestId: %s", *createResponse.OpcWorkRequestId)
		lb.Status.WorkRequestId = createResponse.OpcWorkRequestId
		return lb, lb.Status.HandleError(e)
	}

	lbReq := ocisdklb.GetLoadBalancerRequest{
		LoadBalancerId: lb.Status.Resource.Id,
	}

	lbResp, e := a.lbClient.GetLoadBalancer(a.ctx, lbReq)
	if e != nil {
		glog.Errorf("CreateLoadBalancer GetLoadBalancer error: %v", e)
		return lb, lb.Status.HandleError(err)
	}
	return lb.SetResource(&lbResp.LoadBalancer), lb.Status.HandleError(e)
}

// Delete deletes the load balancer resource in oci
func (a *LoadBalancerAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	var lb = obj.(*ocilbv1alpha1.LoadBalancer)

	deleteRequest := ocisdklb.DeleteLoadBalancerRequest{
		LoadBalancerId: lb.Status.Resource.Id,
	}
	deleteResponse, e := a.lbClient.DeleteLoadBalancer(a.ctx, deleteRequest)
	glog.Infof("DeleteLoadBalancer workrequest: %s", *deleteResponse.OpcWorkRequestId)
	if e == nil && lb.Status.Resource != nil {
		return lb, lb.Status.HandleError(e)
	}
	return lb, e
}

// Get retrieves the load balancer resource from oci
func (a *LoadBalancerAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	var lb = obj.(*ocilbv1alpha1.LoadBalancer)

	if lb.Status.Resource == nil {
		return lb, nil
	}

	request := ocisdklb.GetLoadBalancerRequest{
		LoadBalancerId: lb.Status.Resource.Id,
	}

	r, e := a.lbClient.GetLoadBalancer(a.ctx, request)
	if e != nil {
		return lb, lb.Status.HandleError(e)
	}

	return lb.SetResource(&r.LoadBalancer), lb.Status.HandleError(e)
}

// Update updates the load balancer resource in oci
func (a *LoadBalancerAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	// only attr updateable is displayName which is resource name
	return a.Get(obj)
}

// UpdateForResource calls a common UpdateForResource method to update the load balancer resource in the load balancer object
func (a *LoadBalancerAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
