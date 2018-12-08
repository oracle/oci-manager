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
	"strings"

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
		ocilbv1alpha1.BackendSetKind,
		ocilbv1alpha1.BackendSetResourcePlural,
		ocilbv1alpha1.BackendSetControllerName,
		&ocilbv1alpha1.BackendSetValidation,
		NewBackendSetAdapter)
}

// BackendSetAdapter implements the adapter interface for backend set resource
type BackendSetAdapter struct {
	clientset versioned.Interface
	ctx       context.Context
	lbClient  resourcescommon.LoadBalancerClientInterface
}

// NewBackendSetAdapter creates a new adapter for backend set resource
func NewBackendSetAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	bsa := BackendSetAdapter{}
	bsa.clientset = clientset
	bsa.ctx = context.Background()

	lbClient, err := ocisdklb.NewLoadBalancerClientWithConfigurationProvider(ociconfig)
	if err != nil {
		glog.Errorf("Error creating oci Compute client: %v", err)
		os.Exit(1)
	}
	bsa.lbClient = &lbClient
	return &bsa
}

// Kind returns the resource kind string
func (a *BackendSetAdapter) Kind() string {
	return ocilbv1alpha1.BackendSetKind
}

// Resource returns the plural name of the resource type
func (a *BackendSetAdapter) Resource() string {
	return ocilbv1alpha1.BackendSetResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *BackendSetAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ocilbv1alpha1.SchemeGroupVersion.WithResource(ocilbv1alpha1.BackendSetResourcePlural)
}

// ObjectType returns the backend set type for this adapter
func (a *BackendSetAdapter) ObjectType() runtime.Object {
	return &ocilbv1alpha1.BackendSet{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *BackendSetAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ocilbv1alpha1.BackendSet)
	return ok
}

// Copy returns a copy of a backend set object
func (a *BackendSetAdapter) Copy(obj runtime.Object) runtime.Object {
	BackendSet := obj.(*ocilbv1alpha1.BackendSet)
	return BackendSet.DeepCopyObject()
}

// Equivalent checks if two backend set objects are the same
func (a *BackendSetAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	backendSet1 := obj1.(*ocilbv1alpha1.BackendSet)
	backendSet2 := obj2.(*ocilbv1alpha1.BackendSet)

	return reflect.DeepEqual(backendSet1.Status, backendSet2.Status) &&
		reflect.DeepEqual(backendSet1.Spec, backendSet2.Spec)
}

// IsResourceCompliant checks if resource config is complient with CRD spec
func (a *BackendSetAdapter) IsResourceCompliant(obj runtime.Object) bool {
	backendSet := obj.(*ocilbv1alpha1.BackendSet)
	if backendSet.Status.Resource == nil {
		return false
	}

	return true
}

// IsResourceStatusChanged checks if two objects are the same
func (a *BackendSetAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	return false
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *BackendSetAdapter) Id(obj runtime.Object) string {
	return obj.(*ocilbv1alpha1.BackendSet).GetResourceID()
}

// ObjectMeta returns the object meta struct from the backend set object
func (a *BackendSetAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ocilbv1alpha1.BackendSet).ObjectMeta
}

// DependsOn returns a map of backend set dependencies (objects that the backend set depends on)
func (a *BackendSetAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ocilbv1alpha1.BackendSet).Spec.DependsOn
}

// Dependents returns a map of backend set dependents (objects that depend on the backend set)
func (a *BackendSetAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ocilbv1alpha1.BackendSet).Status.Dependents
}

// CreateObject creates the backend set object
func (a *BackendSetAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocilbv1alpha1.BackendSet)
	return a.clientset.OcilbV1alpha1().BackendSets(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the backend set object
func (a *BackendSetAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocilbv1alpha1.BackendSet)
	return a.clientset.OcilbV1alpha1().BackendSets(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the backend set object
func (a *BackendSetAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var object = obj.(*ocilbv1alpha1.BackendSet)
	return a.clientset.OcilbV1alpha1().BackendSets(object.ObjectMeta.Namespace).Delete(object.Name, options)
}

// DependsOnRefs returns the objects that the backend set depends on
func (a *BackendSetAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	var backendSet = obj.(*ocilbv1alpha1.BackendSet)

	deps := make([]runtime.Object, 0)
	if !resourcescommon.IsOcid(backendSet.Spec.LoadBalancerRef) {
		lb, err := resourcescommon.LoadBalancer(a.clientset, backendSet.ObjectMeta.Namespace, backendSet.Spec.LoadBalancerRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, lb)
	}
	return deps, nil
}

// Create creates the backend set resource in oci
func (a *BackendSetAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	backendSet := obj.(*ocilbv1alpha1.BackendSet)

	if backendSet.Status.WorkRequestId != nil {

		workRequest := ocisdklb.GetWorkRequestRequest{WorkRequestId: backendSet.Status.WorkRequestId}
		workResp, e := a.lbClient.GetWorkRequest(a.ctx, workRequest)
		if e != nil {
			glog.Errorf("CreateBackendSet GetWorkRequest error: %v", e)
			return backendSet, backendSet.Status.HandleError(e)
		}
		glog.Infof("CreateBackendSet workResp state: %s", workResp.LifecycleState)

		if workResp.LifecycleState != ocisdklb.WorkRequestLifecycleStateSucceeded &&
			workResp.LifecycleState != ocisdklb.WorkRequestLifecycleStateFailed {

			if backendSet.Status.WorkRequestStatus == nil ||
				workResp.LifecycleState != *backendSet.Status.WorkRequestStatus {
				backendSet.Status.WorkRequestStatus = &workResp.LifecycleState
				return backendSet, nil
			} else {
				return nil, nil
			}
		}

		if workResp.LifecycleState == ocisdklb.WorkRequestLifecycleStateFailed {
			backendSet.Status.WorkRequestStatus = &workResp.LifecycleState
			err := fmt.Errorf("WorkRequest %s is in failed state", *backendSet.Status.WorkRequestId)
			return backendSet, backendSet.Status.HandleError(err)
		}

		backendSet.Status.WorkRequestId = nil
		backendSet.Status.WorkRequestStatus = nil
	} else {

		if backendSet.Status.LoadBalancerId == nil {
			if resourcescommon.IsOcid(backendSet.Spec.LoadBalancerRef) {
				backendSet.Status.LoadBalancerId = ocisdkcommon.String(backendSet.Spec.LoadBalancerRef)
			} else {
				lbId, err := resourcescommon.LoadBalancerId(a.clientset, backendSet.ObjectMeta.Namespace, backendSet.Spec.LoadBalancerRef)
				if err != nil {
					return backendSet, backendSet.Status.HandleError(err)
				}
				backendSet.Status.LoadBalancerId = &lbId
			}
		}

		beSetReq := ocisdklb.GetBackendSetRequest{
			LoadBalancerId: backendSet.Status.LoadBalancerId,
			BackendSetName: &backendSet.Name,
		}

		r, e := a.lbClient.GetBackendSet(a.ctx, beSetReq)
		if e == nil {
			glog.Infof("CreateBackendSet using existing backendset - reconcile thread faster than create")
			return backendSet.SetResource(&r.BackendSet), nil
		}

		ociHealthChecker := &ocisdklb.HealthCheckerDetails{
			IntervalInMillis:  &backendSet.Spec.HealthChecker.IntervalInMillis,
			Port:              &backendSet.Spec.HealthChecker.Port,
			Protocol:          &backendSet.Spec.HealthChecker.Protocol,
			ResponseBodyRegex: &backendSet.Spec.HealthChecker.ResponseBodyRegex,
			Retries:           &backendSet.Spec.HealthChecker.Retries,
			ReturnCode:        &backendSet.Spec.HealthChecker.ReturnCode,
			TimeoutInMillis:   &backendSet.Spec.HealthChecker.TimeoutInMillis,
			UrlPath:           &backendSet.Spec.HealthChecker.URLPath,
		}

		sslConfig := &ocisdklb.SslConfigurationDetails{}
		if backendSet.Spec.SSLConfig != nil && backendSet.Spec.SSLConfig.CertificateName != "" {
			sslConfig = &ocisdklb.SslConfigurationDetails{
				CertificateName:       &backendSet.Spec.SSLConfig.CertificateName,
				VerifyDepth:           &backendSet.Spec.SSLConfig.VerifyDepth,
				VerifyPeerCertificate: &backendSet.Spec.SSLConfig.VerifyPeerCertificate,
			}
		} else {
			sslConfig = nil
		}

		sessionPersistenceDetails := &ocisdklb.SessionPersistenceConfigurationDetails{}
		if backendSet.Spec.SessionPersistenceConfig != nil && backendSet.Spec.SessionPersistenceConfig.CookieName != "" {
			sessionPersistenceDetails = &ocisdklb.SessionPersistenceConfigurationDetails{
				CookieName:      &backendSet.Spec.SessionPersistenceConfig.CookieName,
				DisableFallback: &backendSet.Spec.SessionPersistenceConfig.DisableFallback,
			}
		} else {
			sessionPersistenceDetails = nil
		}

		createBackendSetDetails := ocisdklb.CreateBackendSetDetails{
			HealthChecker:                   ociHealthChecker,
			Name:                            &backendSet.Name,
			Policy:                          &backendSet.Spec.Policy,
			SessionPersistenceConfiguration: sessionPersistenceDetails,
			SslConfiguration:                sslConfig,
		}

		createBackendSetRequest := ocisdklb.CreateBackendSetRequest{
			CreateBackendSetDetails: createBackendSetDetails,
			LoadBalancerId:          backendSet.Status.LoadBalancerId,
			OpcRetryToken:           ocisdkcommon.String(string(backendSet.UID)),
		}

		createBackendSetResponse, e := a.lbClient.CreateBackendSet(a.ctx, createBackendSetRequest)
		if e != nil {
			glog.Errorf("CreateBackendSet error: %v", e)
			return backendSet, backendSet.Status.HandleError(e)
		}
		glog.Infof("CreateBackendSet workRequestId: %s", *createBackendSetResponse.OpcWorkRequestId)
		backendSet.Status.WorkRequestId = createBackendSetResponse.OpcWorkRequestId
		return backendSet, backendSet.Status.HandleError(e)
	}

	return a.Get(backendSet)
}

// Delete deletes the backend set resource in oci
func (a *BackendSetAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	var backendSet = obj.(*ocilbv1alpha1.BackendSet)

	deleteRequest := ocisdklb.DeleteBackendSetRequest{
		LoadBalancerId: backendSet.Status.LoadBalancerId,
		BackendSetName: &backendSet.Name,
	}
	respMessage, e := a.lbClient.DeleteBackendSet(a.ctx, deleteRequest)
	glog.Infof("DeleteBackendSet resp message: %s", respMessage)

	if e != nil && strings.Contains(e.Error(), "while it's in state DELETING") {
		return backendSet, nil
	}

	return backendSet, backendSet.Status.HandleError(e)
}

// Get retrieves the backend set resource from oci
func (a *BackendSetAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	backendSet := obj.(*ocilbv1alpha1.BackendSet)

	beSetReq := ocisdklb.GetBackendSetRequest{
		LoadBalancerId: backendSet.Status.LoadBalancerId,
		BackendSetName: &backendSet.Name,
	}
	beSetResp, e := a.lbClient.GetBackendSet(a.ctx, beSetReq)
	if e != nil {
		return backendSet, backendSet.Status.HandleError(e)
	}

	return backendSet.SetResource(&beSetResp.BackendSet), backendSet.Status.HandleError(e)
}

// Update updates the backend set resource in oci
func (a *BackendSetAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	backendSet := obj.(*ocilbv1alpha1.BackendSet)

	if backendSet.Status.WorkRequestId != nil {

		workRequest := ocisdklb.GetWorkRequestRequest{WorkRequestId: backendSet.Status.WorkRequestId}
		workResp, e := a.lbClient.GetWorkRequest(a.ctx, workRequest)
		if e != nil {
			glog.Errorf("UpdateBackendSet GetWorkRequest error: %v", e)
			return backendSet, backendSet.Status.HandleError(e)
		}
		glog.Infof("UpdateBackendSet workResp state: %s", workResp.LifecycleState)

		if workResp.LifecycleState != ocisdklb.WorkRequestLifecycleStateSucceeded &&
			workResp.LifecycleState != ocisdklb.WorkRequestLifecycleStateFailed {

			if backendSet.Status.WorkRequestStatus == nil ||
				workResp.LifecycleState != *backendSet.Status.WorkRequestStatus {
				backendSet.Status.WorkRequestStatus = &workResp.LifecycleState
				return backendSet, nil
			} else {
				return nil, nil
			}
		}

		if workResp.LifecycleState == ocisdklb.WorkRequestLifecycleStateFailed {
			backendSet.Status.WorkRequestStatus = &workResp.LifecycleState
			err := fmt.Errorf("WorkRequest %s is in failed state", *backendSet.Status.WorkRequestId)
			return backendSet, backendSet.Status.HandleError(err)
		}

		backendSet.Status.WorkRequestId = nil
		backendSet.Status.WorkRequestStatus = nil

	} else {

		sslConfig := &ocisdklb.SslConfigurationDetails{}
		if backendSet.Spec.SSLConfig != nil && backendSet.Spec.SSLConfig.CertificateName != "" {
			sslConfig = &ocisdklb.SslConfigurationDetails{
				CertificateName:       &backendSet.Spec.SSLConfig.CertificateName,
				VerifyDepth:           &backendSet.Spec.SSLConfig.VerifyDepth,
				VerifyPeerCertificate: &backendSet.Spec.SSLConfig.VerifyPeerCertificate,
			}
		} else {
			sslConfig = nil
		}

		sessionPersistenceDetails := &ocisdklb.SessionPersistenceConfigurationDetails{}
		if backendSet.Spec.SessionPersistenceConfig != nil && backendSet.Spec.SessionPersistenceConfig.CookieName != "" {
			sessionPersistenceDetails = &ocisdklb.SessionPersistenceConfigurationDetails{
				CookieName:      &backendSet.Spec.SessionPersistenceConfig.CookieName,
				DisableFallback: &backendSet.Spec.SessionPersistenceConfig.DisableFallback,
			}
		} else {
			sessionPersistenceDetails = nil
		}

		healthChecker := &ocisdklb.HealthCheckerDetails{
			IntervalInMillis:  &backendSet.Spec.HealthChecker.IntervalInMillis,
			Port:              &backendSet.Spec.HealthChecker.Port,
			Protocol:          &backendSet.Spec.HealthChecker.Protocol,
			ResponseBodyRegex: &backendSet.Spec.HealthChecker.ResponseBodyRegex,
			Retries:           &backendSet.Spec.HealthChecker.Retries,
			ReturnCode:        &backendSet.Spec.HealthChecker.ReturnCode,
			TimeoutInMillis:   &backendSet.Spec.HealthChecker.TimeoutInMillis,
			UrlPath:           &backendSet.Spec.HealthChecker.URLPath,
		}

		updateBackendSetDetails := ocisdklb.UpdateBackendSetDetails{
			Policy:                          &backendSet.Spec.Policy,
			SessionPersistenceConfiguration: sessionPersistenceDetails,
			SslConfiguration:                sslConfig,
			HealthChecker:                   healthChecker,
		}

		updateBackendSetRequest := ocisdklb.UpdateBackendSetRequest{
			BackendSetName:          &backendSet.Name,
			LoadBalancerId:          backendSet.Status.LoadBalancerId,
			UpdateBackendSetDetails: updateBackendSetDetails,
		}

		glog.Infof("update backendset: name: %s, lb id: %s ", backendSet.Name, *backendSet.Status.LoadBalancerId)
		glog.Infof("update backendset: UpdateBackendSetDetails: %v", updateBackendSetDetails)

		updateBackendSetResponse, e := a.lbClient.UpdateBackendSet(a.ctx, updateBackendSetRequest)
		if e != nil {
			glog.Errorf("UpdateBackendSet error: %v", e)
			return backendSet, backendSet.Status.HandleError(e)
		}
		glog.Infof("UpdateBackendSet workRequestId: %s", *updateBackendSetResponse.OpcWorkRequestId)
		backendSet.Status.WorkRequestId = updateBackendSetResponse.OpcWorkRequestId
		return backendSet, backendSet.Status.HandleError(e)
	}

	return a.Get(backendSet)
}

// UpdateForResource calls a common UpdateForResource method to update the backend set resource in the backend set object
func (a *BackendSetAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
