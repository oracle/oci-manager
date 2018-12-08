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
		ocilbv1alpha1.CertificateKind,
		ocilbv1alpha1.CertificateResourcePlural,
		ocilbv1alpha1.CertificateControllerName,
		&ocilbv1alpha1.CertificateValidation,
		NewCertificateAdapter)
}

// CertificateAdapter implements the adapter interface for certificate resource
type CertificateAdapter struct {
	clientset versioned.Interface
	ctx       context.Context
	lbClient  resourcescommon.LoadBalancerClientInterface
}

// NewCertificateAdapter creates a new adapter for certificate resource
func NewCertificateAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	ba := CertificateAdapter{}
	ba.clientset = clientset
	ba.ctx = context.Background()

	lbClient, err := ocisdklb.NewLoadBalancerClientWithConfigurationProvider(ociconfig)
	if err != nil {
		glog.Errorf("Error creating oci LoadBalancer client: %v", err)
		os.Exit(1)
	}
	ba.lbClient = &lbClient
	return &ba
}

// Kind returns the resource kind string
func (a *CertificateAdapter) Kind() string {
	return ocilbv1alpha1.CertificateKind
}

// Resource returns the plural name of the resource type
func (a *CertificateAdapter) Resource() string {
	return ocilbv1alpha1.CertificateResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *CertificateAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ocilbv1alpha1.SchemeGroupVersion.WithResource(ocilbv1alpha1.CertificateResourcePlural)
}

// ObjectType returns the certificate type for this adapter
func (a *CertificateAdapter) ObjectType() runtime.Object {
	return &ocilbv1alpha1.Certificate{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *CertificateAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ocilbv1alpha1.Certificate)
	return ok
}

// Copy returns a copy of a certificate object
func (a *CertificateAdapter) Copy(obj runtime.Object) runtime.Object {
	Certificate := obj.(*ocilbv1alpha1.Certificate)
	return Certificate.DeepCopyObject()
}

// Equivalent checks if two certificate objects are the same
func (a *CertificateAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	Certificate1 := obj1.(*ocilbv1alpha1.Certificate)
	Certificate2 := obj2.(*ocilbv1alpha1.Certificate)
	return reflect.DeepEqual(Certificate1, Certificate2)
}

// IsResourceCompliant checks if resource config is complient with CRD spec
func (a *CertificateAdapter) IsResourceCompliant(obj runtime.Object) bool {
	cert := obj.(*ocilbv1alpha1.Certificate)
	if cert.Status.Resource == nil {
		return false
	}

	if cert.Spec.CACertificate != *cert.Status.Resource.CaCertificate ||
		cert.Spec.PublicCertificate != *cert.Status.Resource.PublicCertificate {
		return false
	}

	return true
}

// IsResourceStatusChanged checks if two objects are the same
func (a *CertificateAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	return false
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *CertificateAdapter) Id(obj runtime.Object) string {
	return obj.(*ocilbv1alpha1.Certificate).GetResourceID()
}

// ObjectMeta returns the object meta struct from the certificate object
func (a *CertificateAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ocilbv1alpha1.Certificate).ObjectMeta
}

// DependsOn returns a map of certificate dependencies (objects that the certificate depends on)
func (a *CertificateAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ocilbv1alpha1.Certificate).Spec.DependsOn
}

// Dependents returns a map of certificate dependents (objects that depend on the certificate)
func (a *CertificateAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ocilbv1alpha1.Certificate).Status.Dependents
}

// CreateObject creates the certificate object
func (a *CertificateAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocilbv1alpha1.Certificate)
	return a.clientset.OcilbV1alpha1().Certificates(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the certificate object
func (a *CertificateAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocilbv1alpha1.Certificate)
	return a.clientset.OcilbV1alpha1().Certificates(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the certificate object
func (a *CertificateAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var object = obj.(*ocilbv1alpha1.Certificate)
	return a.clientset.OcilbV1alpha1().Certificates(object.ObjectMeta.Namespace).Delete(object.Name, options)
}

// DependsOnRefs returns the objects that the certificate depends on
func (a *CertificateAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	var certificate = obj.(*ocilbv1alpha1.Certificate)
	deps := make([]runtime.Object, 0)

	if !resourcescommon.IsOcid(certificate.Spec.LoadBalancerRef) {
		lb, err := resourcescommon.LoadBalancer(a.clientset, certificate.ObjectMeta.Namespace, certificate.Spec.LoadBalancerRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, lb)
	}

	return deps, nil
}

// Create creates the certificate resource in oci
func (a *CertificateAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	certificate := obj.(*ocilbv1alpha1.Certificate)

	if certificate.Status.WorkRequestId != nil {

		workRequest := ocisdklb.GetWorkRequestRequest{WorkRequestId: certificate.Status.WorkRequestId}
		workResp, e := a.lbClient.GetWorkRequest(a.ctx, workRequest)
		if e != nil {
			glog.Errorf("CreateCertificate GetWorkRequest error: %v", e)
			return certificate, certificate.Status.HandleError(e)
		}
		glog.V(4).Infof("CreateCertificate workResp state: %s", workResp.LifecycleState)

		if workResp.LifecycleState != ocisdklb.WorkRequestLifecycleStateSucceeded &&
			workResp.LifecycleState != ocisdklb.WorkRequestLifecycleStateFailed {

			if certificate.Status.WorkRequestStatus == nil ||
				workResp.LifecycleState != *certificate.Status.WorkRequestStatus {
				certificate.Status.WorkRequestStatus = &workResp.LifecycleState
				return certificate, nil
			} else {
				return nil, nil
			}
		}

		if workResp.LifecycleState == ocisdklb.WorkRequestLifecycleStateFailed {
			certificate.Status.WorkRequestStatus = &workResp.LifecycleState
			err := fmt.Errorf("WorkRequest %s is in failed state", *certificate.Status.WorkRequestId)
			return certificate, certificate.Status.HandleError(err)
		}

		certificate.Status.WorkRequestId = nil
		certificate.Status.WorkRequestStatus = nil

	} else {
		if certificate.Status.LoadBalancerId == nil {
			if resourcescommon.IsOcid(certificate.Spec.LoadBalancerRef) {
				certificate.Status.LoadBalancerId = ocisdkcommon.String(certificate.Spec.LoadBalancerRef)
			} else {
				lbId, err := resourcescommon.LoadBalancerId(a.clientset, certificate.ObjectMeta.Namespace, certificate.Spec.LoadBalancerRef)
				if err != nil {
					return certificate, certificate.Status.HandleError(err)
				}
				certificate.Status.LoadBalancerId = &lbId
			}
		}

		// handle when first reconcile mixes with first create
		lbReq := ocisdklb.GetLoadBalancerRequest{
			LoadBalancerId: certificate.Status.LoadBalancerId,
		}
		lbResp, e := a.lbClient.GetLoadBalancer(a.ctx, lbReq)
		r := lbResp.LoadBalancer
		if e == nil && r.Certificates != nil {
			if val, ok := r.Certificates[certificate.Name]; ok {
				glog.Infof("CreateCertificate using existing certificate - reconcile thread faster than create")
				return certificate.SetResource(&val), nil
			}
		}

		createCertDetails := ocisdklb.CreateCertificateDetails{
			CaCertificate:     &certificate.Spec.CACertificate,
			CertificateName:   &certificate.Name,
			Passphrase:        &certificate.Spec.Passphrase,
			PrivateKey:        &certificate.Spec.PrivateKey,
			PublicCertificate: &certificate.Spec.PublicCertificate,
		}

		createCertRequest := ocisdklb.CreateCertificateRequest{
			CreateCertificateDetails: createCertDetails,
			LoadBalancerId:           certificate.Status.LoadBalancerId,
			OpcRetryToken:            ocisdkcommon.String(string(certificate.UID)),
		}

		workResp, e := a.lbClient.CreateCertificate(a.ctx, createCertRequest)
		if e != nil {
			glog.Errorf("CreateCertificate error: %v", e)
			return certificate, certificate.Status.HandleError(e)
		}
		glog.Infof("CreateCertificate workRequestId: %s", *workResp.OpcWorkRequestId)
		certificate.Status.WorkRequestId = workResp.OpcWorkRequestId
		return certificate, certificate.Status.HandleError(e)
	}

	return a.Get(certificate)
}

// Delete deletes the certificate resource in oci
func (a *CertificateAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	certificate := obj.(*ocilbv1alpha1.Certificate)

	deleteRequest := ocisdklb.DeleteCertificateRequest{
		CertificateName: &certificate.Name,
		LoadBalancerId:  certificate.Status.LoadBalancerId,
	}
	respMessage, e := a.lbClient.DeleteCertificate(a.ctx, deleteRequest)
	glog.Infof("resp message: %s", respMessage)

	return certificate, certificate.Status.HandleError(e)
}

// Get retrieves the certificate resource from oci
func (a *CertificateAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	certificate := obj.(*ocilbv1alpha1.Certificate)

	lbReq := ocisdklb.GetLoadBalancerRequest{
		LoadBalancerId: certificate.Status.LoadBalancerId,
	}

	lbResp, e := a.lbClient.GetLoadBalancer(a.ctx, lbReq)
	r := lbResp.LoadBalancer
	if e == nil && r.Certificates != nil {
		if val, ok := r.Certificates[certificate.Name]; ok {
			certificate.SetResource(&val)
		}
	}

	return certificate, certificate.Status.HandleError(e)
}

// Update updates the certificate resource in oci
func (a *CertificateAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	certificate := obj.(*ocilbv1alpha1.Certificate)
	// no update certificate api
	// TODO: (for update cert) add logic to remove cert from listener, delete and add cert, then add to listener
	return certificate, certificate.Status.HandleError(nil)
}

// UpdateForResource calls a common UpdateForResource method to update the certificate resource in the certificate object
func (a *CertificateAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
