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
	"strconv"

	ocisdkcommon "github.com/oracle/oci-go-sdk/common"
	ocisdkcore "github.com/oracle/oci-go-sdk/core"
	ocisdklb "github.com/oracle/oci-go-sdk/loadbalancer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"errors"

	ocicommon "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	lbgroup "github.com/oracle/oci-manager/pkg/apis/ocilb.oracle.com"
	ocilbv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocilb.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
	"strings"
)

func init() {
	resourcescommon.RegisterResourceTypeWithValidation(
		lbgroup.GroupName,
		ocilbv1alpha1.BackendKind,
		ocilbv1alpha1.BackendResourcePlural,
		ocilbv1alpha1.BackendControllerName,
		&ocilbv1alpha1.BackendValidation,
		NewBackendAdapter)
}

// BackendAdapter implements the adapter interface for backend resource
type BackendAdapter struct {
	clientset versioned.Interface
	ctx       context.Context
	cClient   resourcescommon.ComputeClientInterface
	lbClient  resourcescommon.LoadBalancerClientInterface
	vcnClient resourcescommon.VcnClientInterface
}

// NewBackendAdapter creates a new adapter for backend resource
func NewBackendAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	ba := BackendAdapter{}
	ba.clientset = clientset
	ba.ctx = context.Background()

	lbClient, err := ocisdklb.NewLoadBalancerClientWithConfigurationProvider(ociconfig)
	if err != nil {
		glog.Errorf("Error creating oci LoadBalancer client: %v", err)
		os.Exit(1)
	}
	ba.lbClient = &lbClient

	cClient, err := ocisdkcore.NewComputeClientWithConfigurationProvider(ociconfig)
	if err != nil {
		glog.Errorf("Error creating oci Compute client: %v", err)
		os.Exit(1)
	}
	ba.cClient = &cClient

	vcnClient, err := ocisdkcore.NewVirtualNetworkClientWithConfigurationProvider(ociconfig)
	if err != nil {
		glog.Errorf("Error creating oci VCN client: %v", err)
		os.Exit(1)
	}
	ba.vcnClient = &vcnClient
	return &ba
}

// Kind returns the resource kind string
func (a *BackendAdapter) Kind() string {
	return ocilbv1alpha1.BackendKind
}

// Resource returns the plural name of the resource type
func (a *BackendAdapter) Resource() string {
	return ocilbv1alpha1.BackendResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *BackendAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ocilbv1alpha1.SchemeGroupVersion.WithResource(ocilbv1alpha1.BackendResourcePlural)
}

// ObjectType returns the backend type for this adapter
func (a *BackendAdapter) ObjectType() runtime.Object {
	return &ocilbv1alpha1.Backend{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *BackendAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ocilbv1alpha1.Backend)
	return ok
}

// Copy returns a copy of a backend object
func (a *BackendAdapter) Copy(obj runtime.Object) runtime.Object {
	Backend := obj.(*ocilbv1alpha1.Backend)
	return Backend.DeepCopyObject()
}

// Equivalent checks if two backend objects are the same
func (a *BackendAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	backend1 := obj1.(*ocilbv1alpha1.Backend)
	backend2 := obj2.(*ocilbv1alpha1.Backend)
	if backend1.Status.Resource != nil {
		backend1.Spec.Weight = *backend1.Status.Resource.Weight
		backend1.Spec.Port = *backend1.Status.Resource.Port
	}
	return reflect.DeepEqual(backend1, backend2)
}

// IsResourceCompliant checks if resource config is complient with CRD spec
func (a *BackendAdapter) IsResourceCompliant(obj runtime.Object) bool {
	backend := obj.(*ocilbv1alpha1.Backend)

	if backend.Status.Resource == nil {
		return false
	}

	if backend.Spec.Weight != *backend.Status.Resource.Weight ||
		backend.Spec.Port != *backend.Status.Resource.Port {
		return false
	}

	return true
}

// IsResourceStatusChanged checks if two vcn objects are the same
func (a *BackendAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	return false
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *BackendAdapter) Id(obj runtime.Object) string {
	return obj.(*ocilbv1alpha1.Backend).GetResourceID()
}

// ObjectMeta returns the object meta struct from the backend object
func (a *BackendAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ocilbv1alpha1.Backend).ObjectMeta
}

// DependsOn returns a map of backend dependencies (objects that the backend depends on)
func (a *BackendAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ocilbv1alpha1.Backend).Spec.DependsOn
}

// Dependents returns a map of backend dependents (objects that depend on the backend)
func (a *BackendAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ocilbv1alpha1.Backend).Status.Dependents
}

// CreateObject creates the backend object
func (a *BackendAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocilbv1alpha1.Backend)
	return a.clientset.OcilbV1alpha1().Backends(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the backend object
func (a *BackendAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocilbv1alpha1.Backend)
	return a.clientset.OcilbV1alpha1().Backends(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the backend object
func (a *BackendAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var be = obj.(*ocilbv1alpha1.Backend)
	return a.clientset.OcilbV1alpha1().Backends(be.ObjectMeta.Namespace).Delete(be.Name, options)
}

// DependsOnRefs returns the objects that the backend depends on
func (a *BackendAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	var backend = obj.(*ocilbv1alpha1.Backend)
	deps := make([]runtime.Object, 0)

	if !resourcescommon.IsOcid(backend.Spec.LoadBalancerRef) {
		lb, err := resourcescommon.LoadBalancer(a.clientset, backend.ObjectMeta.Namespace, backend.Spec.LoadBalancerRef)
		if err != nil {
			glog.Errorf("Backend DependsOnRefs lb err: %v", err)
			return nil, err
		}
		deps = append(deps, lb)
	}

	bs, err := resourcescommon.BackendSet(a.clientset, backend.ObjectMeta.Namespace, backend.Spec.BackendSetRef)
	if err != nil {
		glog.Errorf("Backend DependsOnRefs backendset err: %v", err)
		return nil, err
	}
	deps = append(deps, bs)

	instance, err := resourcescommon.Instance(a.clientset, backend.ObjectMeta.Namespace, backend.Spec.InstanceRef)
	if err != nil {
		glog.Errorf("Backend DependsOnRefs instance err: %v", err)
		return nil, err
	}
	deps = append(deps, instance)
	return deps, nil
}

func getBackendName(backend *ocilbv1alpha1.Backend) *string {
	backendName := backend.Spec.IPAddress + ":" + strconv.Itoa(backend.Spec.Port)
	return &backendName
}

// Create creates the backend resource in oci
func (a *BackendAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	backend := obj.(*ocilbv1alpha1.Backend)

	if backend.Status.WorkRequestId != nil {

		workRequest := ocisdklb.GetWorkRequestRequest{WorkRequestId: backend.Status.WorkRequestId}
		workResp, e := a.lbClient.GetWorkRequest(a.ctx, workRequest)
		if e != nil {
			glog.Errorf("CreateBackend GetWorkRequest error: %v", e)
			return backend, backend.Status.HandleError(e)
		}
		glog.Infof("CreateBackend workResp state: %s", workResp.LifecycleState)

		if workResp.LifecycleState != ocisdklb.WorkRequestLifecycleStateSucceeded &&
			workResp.LifecycleState != ocisdklb.WorkRequestLifecycleStateFailed {

			if backend.Status.WorkRequestStatus == nil ||
				workResp.LifecycleState != *backend.Status.WorkRequestStatus {
				backend.Status.WorkRequestStatus = &workResp.LifecycleState
				return backend, nil
			} else {
				return nil, nil
			}
		}

		if workResp.LifecycleState == ocisdklb.WorkRequestLifecycleStateFailed {
			backend.Status.WorkRequestStatus = &workResp.LifecycleState
			err := fmt.Errorf("WorkRequest %s is in failed state", *backend.Status.WorkRequestId)
			return backend, backend.Status.HandleError(err)
		}

		backend.Status.WorkRequestId = nil
		backend.Status.WorkRequestStatus = nil

	} else {

		if backend.Status.LoadBalancerId == nil {
			if resourcescommon.IsOcid(backend.Spec.LoadBalancerRef) {
				backend.Status.LoadBalancerId = ocisdkcommon.String(backend.Spec.LoadBalancerRef)
			} else {
				lbId, err := resourcescommon.LoadBalancerId(a.clientset, backend.ObjectMeta.Namespace, backend.Spec.LoadBalancerRef)
				if err != nil {
					return backend, backend.Status.HandleError(err)
				}
				backend.Status.LoadBalancerId = &lbId
			}
		}

		instance, err := resourcescommon.Instance(a.clientset, backend.ObjectMeta.Namespace, backend.Spec.InstanceRef)
		if err != nil {
			return backend, backend.Status.HandleError(err)
		}

		if instance.Status.Resource == nil || instance.Status.PrimaryVnic == nil || *instance.Status.Resource.Id == "" {
			return backend, backend.Status.HandleError(errors.New("Instance resource is not created"))
		}

		glog.Infof("CreateBackend instance primary vnic private ip: %s", *instance.Status.PrimaryVnic.PrivateIp)
		backend.Spec.IPAddress = *instance.Status.PrimaryVnic.PrivateIp

		backendName := getBackendName(backend)

		// handle when first reconcile overlaps with first create
		getBackendReq := ocisdklb.GetBackendRequest{
			BackendName:    backendName,
			BackendSetName: &backend.Spec.BackendSetRef,
			LoadBalancerId: backend.Status.LoadBalancerId,
		}
		backendResp, e := a.lbClient.GetBackend(a.ctx, getBackendReq)
		if e == nil {
			glog.Infof("CreateBackend using existing backend - reconcile thread faster than create")
			return backend.SetResource(&backendResp.Backend), nil
		}

		details := ocisdklb.CreateBackendDetails{
			Backup:    &backend.Spec.Backup,
			Drain:     &backend.Spec.Drain,
			IpAddress: &backend.Spec.IPAddress,
			Port:      &backend.Spec.Port,
			Offline:   &backend.Spec.Offline,
			Weight:    &backend.Spec.Weight,
		}

		createRequest := ocisdklb.CreateBackendRequest{
			BackendSetName:       &backend.Spec.BackendSetRef,
			CreateBackendDetails: details,
			LoadBalancerId:       backend.Status.LoadBalancerId,
			OpcRetryToken:        ocisdkcommon.String(string(backend.UID)),
		}

		glog.Infof("Backend: %s OpcRetryToken: %s", *backendName, string(backend.UID))

		createResponse, e := a.lbClient.CreateBackend(a.ctx, createRequest)
		if e != nil {
			glog.Errorf("CreateBackend error: %v", e)
			return backend, backend.Status.HandleError(e)
		}

		glog.Infof("CreateBackend workRequestId: %s", *createResponse.OpcWorkRequestId)
		backend.Status.WorkRequestId = createResponse.OpcWorkRequestId
		return backend, backend.Status.HandleError(e)
	}

	return a.Get(backend)
}

// Delete deletes the backend resource in oci
func (a *BackendAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	var be = obj.(*ocilbv1alpha1.Backend)

	deleteRequest := ocisdklb.DeleteBackendRequest{
		LoadBalancerId: be.Status.LoadBalancerId,
		BackendSetName: &be.Spec.BackendSetRef,
		BackendName:    be.Status.Resource.Name,
	}
	respMessage, err := a.lbClient.DeleteBackend(a.ctx, deleteRequest)
	if err != nil {
		glog.Infof("DeleteBackend name: %s error: %v", be.Name, err)
		if strings.Contains(err.Error(), "while it's in state DELETING") {
			return be, nil
		}
		return be, be.Status.HandleError(err)

	}
	glog.Infof("DeleteBackend resp message: %s", respMessage)
	return be, nil
}

// Get retrieves the backend resource from oci
func (a *BackendAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	backend := obj.(*ocilbv1alpha1.Backend)

	getBackendReq := ocisdklb.GetBackendRequest{
		BackendName:    getBackendName(backend),
		BackendSetName: &backend.Spec.BackendSetRef,
		LoadBalancerId: backend.Status.LoadBalancerId,
	}
	backendResp, e := a.lbClient.GetBackend(a.ctx, getBackendReq)
	if e != nil {
		return backend, backend.Status.HandleError(e)
	}

	return backend.SetResource(&backendResp.Backend), backend.Status.HandleError(e)
}

// Update updates the backend resource in oci
func (a *BackendAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	backend := obj.(*ocilbv1alpha1.Backend)

	if backend.Status.WorkRequestId != nil {

		workRequest := ocisdklb.GetWorkRequestRequest{WorkRequestId: backend.Status.WorkRequestId}
		workResp, e := a.lbClient.GetWorkRequest(a.ctx, workRequest)
		if e != nil {
			glog.Errorf("UpdateBackend GetWorkRequest error: %v", e)
			return backend, backend.Status.HandleError(e)
		}
		glog.Infof("UpdateBackend workResp state: %s", workResp.LifecycleState)

		if workResp.LifecycleState != ocisdklb.WorkRequestLifecycleStateSucceeded &&
			workResp.LifecycleState != ocisdklb.WorkRequestLifecycleStateFailed {

			if backend.Status.WorkRequestStatus == nil ||
				workResp.LifecycleState != *backend.Status.WorkRequestStatus {
				backend.Status.WorkRequestStatus = &workResp.LifecycleState
				return backend, nil
			} else {
				return nil, nil
			}
		}

		if workResp.LifecycleState == ocisdklb.WorkRequestLifecycleStateFailed {
			backend.Status.WorkRequestStatus = &workResp.LifecycleState
			err := fmt.Errorf("WorkRequest %s is in failed state", *backend.Status.WorkRequestId)
			return backend, backend.Status.HandleError(err)
		}

		backend.Status.WorkRequestId = nil
		backend.Status.WorkRequestStatus = nil

	} else {
		details := ocisdklb.UpdateBackendDetails{
			Backup:  &backend.Spec.Backup,
			Drain:   &backend.Spec.Drain,
			Offline: &backend.Spec.Offline,
			Weight:  &backend.Spec.Weight,
		}

		updateRequest := ocisdklb.UpdateBackendRequest{
			BackendName:          getBackendName(backend),
			BackendSetName:       &backend.Spec.BackendSetRef,
			UpdateBackendDetails: details,
			LoadBalancerId:       backend.Status.LoadBalancerId,
		}

		updateResponse, e := a.lbClient.UpdateBackend(a.ctx, updateRequest)
		if e != nil {
			glog.Errorf("UpdateBackend error: %v", e)
			return backend, backend.Status.HandleError(e)
		}
		glog.Infof("UpdateBackend workRequestId: %s", *updateResponse.OpcWorkRequestId)
		backend.Status.WorkRequestId = updateResponse.OpcWorkRequestId
		return backend, backend.Status.HandleError(e)
	}

	return a.Get(backend)
}

// UpdateForResource calls a common UpdateForResource method to update the backend resource in the backend object
func (a *BackendAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
