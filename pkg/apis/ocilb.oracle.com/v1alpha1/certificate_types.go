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

package v1alpha1

import (
	ocilb "github.com/oracle/oci-go-sdk/loadbalancer"
	common "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Certificate names
const (
	CertificateKind           = "Certificate"
	CertificateResourcePlural = "certificates"
	CertificateControllerName = "certificates"
)

// CertificateValidation describes the certificate validation schema
var CertificateValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"loadBalancerRef", "publicCertificate", "privateKey"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"loadBalancerRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},

					"publicCertificate": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"privateKey": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"caCertificate": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"passphrase": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
				},
			},
		},
	},
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Certificate describes a certificate
type Certificate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              CertificateSpec   `json:"spec"`
	Status            CertificateStatus `json:"status,omitempty"`
}

// CertificateSpec describes a certificate spec
type CertificateSpec struct {
	LoadBalancerRef string `json:"loadBalancerRef"`

	PublicCertificate string `header:"-" url:"-" json:"publicCertificate"`
	PrivateKey        string `header:"-" url:"-" json:"privateKey,omitempty"` // Only for create

	// Optional
	CACertificate string `header:"-" url:"-" json:"caCertificate,omitempty"`
	Passphrase    string `header:"-" url:"-" json:"passphrase,omitempty"` // Only for create

	common.Dependency
}

// CertificateStatus describes a certificate status
type CertificateStatus struct {
	common.ResourceStatus
	LoadBalancerId *string

	WorkRequestId     *string                              `json:"workRequestId,omitempty"`
	WorkRequestStatus *ocilb.WorkRequestLifecycleStateEnum `json:"workRequestStatus,omitempty"`

	Resource *CertificateResource `json:"resource,omitempty"`
}

// CertificateResource describes a certificate resource from oci
type CertificateResource struct {
	*ocilb.Certificate
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertificateList is a list of Certificate items
type CertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Certificate `json:"items"`
}

// IsResource returns true if there is an oci id, otherwise false
func (s *Certificate) IsResource() bool {
	if s.GetResourceID() == "" {
		return false
	}
	return true
}

// GetResourceID returns the oci id of the certificate
func (s *Certificate) GetResourceID() string {
	var id string
	if s.Status.Resource != nil {
		id = *s.Status.Resource.CertificateName
	}
	return id
}

// GetResourcePlural returns the plural name of the certificate type
func (s *Certificate) GetResourcePlural() string {
	return CertificateResourcePlural
}

// GetGroupVersionResource returns the group version of the certificate type
func (s *Certificate) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(CertificateResourcePlural)
}

// SetResource sets the resource in the certificate status
func (s *Certificate) SetResource(r *ocilb.Certificate) *Certificate {
	if r != nil {
		s.Status.Resource = &CertificateResource{r}
	}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *Certificate) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds a certificate dependent
func (s *Certificate) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes a certificate dependent
func (s *Certificate) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the certificate dependent is registered
func (s *Certificate) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the certificate spec
func (in *CertificateSpec) DeepCopy() *CertificateSpec {
	if in == nil {
		return nil
	}
	out := in
	return out
}

// DeepCopy the certificate oci resource
func (in *CertificateResource) DeepCopy() (out *CertificateResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
