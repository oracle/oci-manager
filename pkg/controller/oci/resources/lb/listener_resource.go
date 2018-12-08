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
		ocilbv1alpha1.ListenerKind,
		ocilbv1alpha1.ListenerResourcePlural,
		ocilbv1alpha1.ListenerControllerName,
		&ocilbv1alpha1.ListenerValidation,
		NewListenerAdapter)
}

// ListenerAdapter implements the adapter interface for listener resource
type ListenerAdapter struct {
	clientset versioned.Interface
	ctx       context.Context
	lbClient  resourcescommon.LoadBalancerClientInterface
}

// NewListenerAdapter creates a new adapter for listener resource
func NewListenerAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	la := ListenerAdapter{}
	la.clientset = clientset
	la.ctx = context.Background()

	lbClient, err := ocisdklb.NewLoadBalancerClientWithConfigurationProvider(ociconfig)
	if err != nil {
		glog.Errorf("Error creating oci LoadBalancer client: %v", err)
		os.Exit(1)
	}
	la.lbClient = &lbClient
	return &la
}

// Kind returns the resource kind string
func (a *ListenerAdapter) Kind() string {
	return ocilbv1alpha1.ListenerKind
}

// Resource returns the plural name of the resource type
func (a *ListenerAdapter) Resource() string {
	return ocilbv1alpha1.ListenerResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *ListenerAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ocilbv1alpha1.SchemeGroupVersion.WithResource(ocilbv1alpha1.ListenerResourcePlural)
}

// ObjectType returns the listener type for this adapter
func (a *ListenerAdapter) ObjectType() runtime.Object {
	return &ocilbv1alpha1.Listener{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *ListenerAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ocilbv1alpha1.Listener)
	return ok
}

// Copy returns a copy of a listener object
func (a *ListenerAdapter) Copy(obj runtime.Object) runtime.Object {
	Listener := obj.(*ocilbv1alpha1.Listener)
	return Listener.DeepCopyObject()
}

// Equivalent checks if two listener objects are the same
func (a *ListenerAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	listener1 := obj1.(*ocilbv1alpha1.Listener)
	listener2 := obj2.(*ocilbv1alpha1.Listener)
	if listener1.Status.Resource != nil {
		if listener1.Status.Resource.SslConfiguration == nil {
			listener1.Spec.CertificateRef = ""
		} else {
			listener1.Spec.CertificateRef = *listener1.Status.Resource.SslConfiguration.CertificateName
		}
		listener1.Spec.Port = *listener1.Status.Resource.Port
		listener1.Spec.Protocol = *listener1.Status.Resource.Protocol
	}
	return reflect.DeepEqual(listener1, listener2)
}

// IsResourceCompliant checks if resource config is complient with CRD spec
func (a *ListenerAdapter) IsResourceCompliant(obj runtime.Object) bool {
	listener := obj.(*ocilbv1alpha1.Listener)

	if listener.Status.WorkRequestId != nil {
		return false
	}

	if listener.Status.Resource == nil {
		return false
	}

	if listener.Spec.Port != *listener.Status.Resource.Port ||
		listener.Spec.Protocol != *listener.Status.Resource.Protocol {
		return true
	}

	return true
}

// IsResourceStatusChanged checks if two vcn objects are the same
func (a *ListenerAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	return false
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *ListenerAdapter) Id(obj runtime.Object) string {
	return obj.(*ocilbv1alpha1.Listener).GetResourceID()
}

// ObjectMeta returns the object meta struct from the listener object
func (a *ListenerAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ocilbv1alpha1.Listener).ObjectMeta
}

// DependsOn returns a map of listener dependencies (objects that the listener depends on)
func (a *ListenerAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ocilbv1alpha1.Listener).Spec.DependsOn
}

// Dependents returns a map of listener dependents (objects that depend on the listener)
func (a *ListenerAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ocilbv1alpha1.Listener).Status.Dependents
}

// CreateObject creates the listener object
func (a *ListenerAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocilbv1alpha1.Listener)
	return a.clientset.OcilbV1alpha1().Listeners(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the listener object
func (a *ListenerAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocilbv1alpha1.Listener)
	return a.clientset.OcilbV1alpha1().Listeners(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the listener object
func (a *ListenerAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var object = obj.(*ocilbv1alpha1.Listener)
	return a.clientset.OcilbV1alpha1().Listeners(object.ObjectMeta.Namespace).Delete(object.Name, options)
}

// DependsOnRefs returns the objects that the listener depends on
func (a *ListenerAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	var listener = obj.(*ocilbv1alpha1.Listener)

	deps := make([]runtime.Object, 0)

	if !resourcescommon.IsOcid(listener.Spec.LoadBalancerRef) {
		lb, err := resourcescommon.LoadBalancer(a.clientset, listener.ObjectMeta.Namespace, listener.Spec.LoadBalancerRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, lb)
	}
	return deps, nil
}

// Create creates the listener resource in oci
func (a *ListenerAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	listener := obj.(*ocilbv1alpha1.Listener)

	if listener.Status.WorkRequestId != nil {

		workRequest := ocisdklb.GetWorkRequestRequest{WorkRequestId: listener.Status.WorkRequestId}
		workResp, e := a.lbClient.GetWorkRequest(a.ctx, workRequest)
		if e != nil {
			glog.Errorf("CreateListener GetWorkRequest error: %v", e)
			return listener, listener.Status.HandleError(e)
		}
		glog.Infof("CreateListener workResp state: %s", workResp.LifecycleState)

		if workResp.LifecycleState != ocisdklb.WorkRequestLifecycleStateSucceeded &&
			workResp.LifecycleState != ocisdklb.WorkRequestLifecycleStateFailed {

			if listener.Status.WorkRequestStatus == nil ||
				workResp.LifecycleState != *listener.Status.WorkRequestStatus {
				listener.Status.WorkRequestStatus = &workResp.LifecycleState
				return listener, nil
			} else {
				return nil, nil
			}
		}

		if workResp.LifecycleState == ocisdklb.WorkRequestLifecycleStateFailed {
			listener.Status.WorkRequestStatus = &workResp.LifecycleState
			err := fmt.Errorf("WorkRequest %s is in failed state", *listener.Status.WorkRequestId)
			return listener, listener.Status.HandleError(err)
		}

		listener.Status.WorkRequestId = nil
		listener.Status.WorkRequestStatus = nil

	} else {
		if listener.Status.LoadBalancerId == nil {
			if resourcescommon.IsOcid(listener.Spec.LoadBalancerRef) {
				listener.Status.LoadBalancerId = ocisdkcommon.String(listener.Spec.LoadBalancerRef)
			} else {
				lbId, err := resourcescommon.LoadBalancerId(a.clientset, listener.ObjectMeta.Namespace, listener.Spec.LoadBalancerRef)
				if err != nil {
					return listener, listener.Status.HandleError(err)
				}
				listener.Status.LoadBalancerId = ocisdkcommon.String(lbId)
			}
		}

		backendSet, e := resourcescommon.BackendSet(a.clientset, listener.ObjectMeta.Namespace, listener.Spec.DefaultBackendSetName)
		if e != nil {
			return listener, listener.Status.HandleError(e)
		}
		if backendSet.Status.Resource != nil && *backendSet.Status.Resource.Name != "" {
			glog.V(4).Infof("CreateListener backendset: %s has status: %s", listener.Spec.DefaultBackendSetName, backendSet.Status.State)
		} else {
			return listener, listener.Status.HandleError(errors.New("BackendSet resource is not created"))
		}

		lbReq := ocisdklb.GetLoadBalancerRequest{
			LoadBalancerId: listener.Status.LoadBalancerId,
		}

		lbResp, e := a.lbClient.GetLoadBalancer(a.ctx, lbReq)
		if e == nil && lbResp.LoadBalancer.Listeners != nil {
			if val, ok := lbResp.LoadBalancer.Listeners[listener.Name]; ok {
				glog.Infof("using existing listener in create - reconcile loop faster that dependency creations")
				return listener.SetResource(&val), nil
			}
		}

		sslConfig := &ocisdklb.SslConfigurationDetails{}
		if listener.Spec.CertificateRef != "" {
			sslConfig.CertificateName = &listener.Spec.CertificateRef
		} else {
			sslConfig = nil
		}

		connectionConfig := &ocisdklb.ConnectionConfiguration{
			IdleTimeout: &listener.Spec.IdleTimeout,
		}
		createListenerDetails := ocisdklb.CreateListenerDetails{
			ConnectionConfiguration: connectionConfig,
			DefaultBackendSetName:   &listener.Spec.DefaultBackendSetName,
			Name:                    &listener.Name,
			Port:                    &listener.Spec.Port,
			Protocol:                &listener.Spec.Protocol,
			SslConfiguration:        sslConfig,
		}

		if listener.Spec.PathRouteSetName != "" {
			createListenerDetails.PathRouteSetName = &listener.Spec.PathRouteSetName
		}

		createListenerRequest := ocisdklb.CreateListenerRequest{
			CreateListenerDetails: createListenerDetails,
			LoadBalancerId:        listener.Status.LoadBalancerId,
			OpcRetryToken:         ocisdkcommon.String(string(listener.UID)),
		}

		createListenerResp, e := a.lbClient.CreateListener(a.ctx, createListenerRequest)
		if e != nil {
			glog.Errorf("CreateListener error: %v", e)
			return listener, listener.Status.HandleError(e)
		}

		glog.Infof("CreateListener workRequestId: %s", *createListenerResp.OpcWorkRequestId)
		listener.Status.WorkRequestId = createListenerResp.OpcWorkRequestId
		return listener, listener.Status.HandleError(e)
	}

	lbReq := ocisdklb.GetLoadBalancerRequest{
		LoadBalancerId: listener.Status.LoadBalancerId,
	}
	lbResp, e := a.lbClient.GetLoadBalancer(a.ctx, lbReq)
	if e == nil && lbResp.LoadBalancer.Listeners != nil {
		if val, ok := lbResp.LoadBalancer.Listeners[listener.Name]; ok {
			return listener.SetResource(&val), nil
		}
	}

	return listener, listener.Status.HandleError(e)
}

// Delete deletes the listener resource in oci
func (a *ListenerAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	listener := obj.(*ocilbv1alpha1.Listener)

	deleteReq := ocisdklb.DeleteListenerRequest{
		LoadBalancerId: listener.Status.LoadBalancerId,
		ListenerName:   &listener.Name,
	}
	respMessage, e := a.lbClient.DeleteListener(a.ctx, deleteReq)
	glog.Infof("DeleteListener response message: %s", respMessage)

	if e != nil && strings.Contains(e.Error(), "not found") {
		e = nil
	}
	return listener, listener.Status.HandleError(e)
}

// Get retrieves the listener resource from oci
func (a *ListenerAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	var listener = obj.(*ocilbv1alpha1.Listener)
	return listener, listener.Status.HandleError(nil)
}

// Update updates the listener resource in oci
func (a *ListenerAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	listener := obj.(*ocilbv1alpha1.Listener)

	if listener.Status.WorkRequestId != nil {

		workRequest := ocisdklb.GetWorkRequestRequest{WorkRequestId: listener.Status.WorkRequestId}
		workResp, e := a.lbClient.GetWorkRequest(a.ctx, workRequest)
		if e != nil {
			glog.Errorf("UpdateListener GetWorkRequest error: %v", e)
			return listener, listener.Status.HandleError(e)
		}
		glog.Infof("UpdateListener workResp state: %s", workResp.LifecycleState)

		if workResp.LifecycleState != ocisdklb.WorkRequestLifecycleStateSucceeded &&
			workResp.LifecycleState != ocisdklb.WorkRequestLifecycleStateFailed {

			if listener.Status.WorkRequestStatus == nil ||
				workResp.LifecycleState != *listener.Status.WorkRequestStatus {
				listener.Status.WorkRequestStatus = &workResp.LifecycleState
				return listener, nil
			} else {
				return nil, nil
			}
		}

		if workResp.LifecycleState == ocisdklb.WorkRequestLifecycleStateFailed {
			listener.Status.WorkRequestStatus = &workResp.LifecycleState
			err := fmt.Errorf("WorkRequest %s is in failed state", *listener.Status.WorkRequestId)
			return listener, listener.Status.HandleError(err)
		}

		listener.Status.WorkRequestId = nil
		listener.Status.WorkRequestStatus = nil

	} else {

		sslConfig := &ocisdklb.SslConfigurationDetails{}
		if listener.Spec.CertificateRef != "" {
			sslConfig.CertificateName = &listener.Spec.CertificateRef
		} else {
			sslConfig = nil
		}

		connectionConfig := &ocisdklb.ConnectionConfiguration{
			IdleTimeout: &listener.Spec.IdleTimeout,
		}

		updateListenerDetails := ocisdklb.UpdateListenerDetails{
			ConnectionConfiguration: connectionConfig,
			DefaultBackendSetName:   &listener.Spec.DefaultBackendSetName,
			Port:                    &listener.Spec.Port,
			Protocol:                &listener.Spec.Protocol,
			SslConfiguration:        sslConfig,
		}

		if listener.Spec.PathRouteSetName != "" {
			updateListenerDetails.PathRouteSetName = &listener.Spec.PathRouteSetName
		}

		updateListenerReq := ocisdklb.UpdateListenerRequest{
			ListenerName:          &listener.Name,
			LoadBalancerId:        listener.Status.LoadBalancerId,
			UpdateListenerDetails: updateListenerDetails,
		}

		updateListenerResp, e := a.lbClient.UpdateListener(a.ctx, updateListenerReq)
		if e != nil {
			glog.Errorf("UpdateListener error: %v", e)
			return listener, listener.Status.HandleError(e)
		}

		glog.Infof("UpdateListener workRequestId: %s", *updateListenerResp.OpcWorkRequestId)
		listener.Status.WorkRequestId = updateListenerResp.OpcWorkRequestId
		return listener, listener.Status.HandleError(e)
	}

	lbReq := ocisdklb.GetLoadBalancerRequest{
		LoadBalancerId: listener.Status.LoadBalancerId,
	}

	lbResp, e := a.lbClient.GetLoadBalancer(a.ctx, lbReq)
	if e == nil && lbResp.LoadBalancer.Listeners != nil {
		if val, ok := lbResp.LoadBalancer.Listeners[listener.Name]; ok {
			return listener.SetResource(&val), nil
		}
	}

	return listener, listener.Status.HandleError(e)
}

// UpdateForResource calls a common UpdateForResource method to update the listener resource in the listener object
func (a *ListenerAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
