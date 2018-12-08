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

// BackendSet names
const (
	BackendSetKind           = "BackendSet"
	BackendSetResourcePlural = "backendsets"
	BackendSetControllerName = "backendsets"
)

var oneSecMillis = float64(1000)
var oneDayMillis = float64(86400000)

var minReturnCode = float64(100)
var maxReturnCode = float64(600)

var minDepth = float64(1)
var maxDepth = float64(10)

// BackendSetValidation describes the backend set validation schema
var BackendSetValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"loadBalancerRef"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"loadBalancerRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"policy": {
						Type:    common.ValidationTypeString,
						Pattern: "^ROUND_ROBIN$|^LEAST_CONNECTIONS$|^IP_HASH$",
					},
					"healthChecker": {
						Properties: map[string]apiextv1beta1.JSONSchemaProps{
							"intervalInMillis": {
								Type:    common.ValidationTypeInteger,
								Minimum: &oneSecMillis,
								// TODO: get proper max
								Maximum: &oneDayMillis,
							},
							"port": {
								Type:    common.ValidationTypeInteger,
								Minimum: &minTCPPort,
								Maximum: &maxTCPPort,
							},
							"protocol": {
								Type:    common.ValidationTypeString,
								Pattern: common.LoadBalancerProtocolRegex,
							},
							"responseBodyRegex": {
								Type:    common.ValidationTypeString,
								Pattern: common.AnyStringValidationRegex,
							},
							"retries": {
								Type: common.ValidationTypeInteger,
							},
							"returnCode": {
								Type:    common.ValidationTypeInteger,
								Minimum: &minReturnCode,
								Maximum: &maxReturnCode,
							},
							"timeoutInMillis": {
								Type:    common.ValidationTypeInteger,
								Minimum: &oneSecMillis,
								// TODO: get proper max
								Maximum: &oneDayMillis,
							},
							"urlPath": {
								Type:    common.ValidationTypeString,
								Pattern: "^\\/$|^\\/[a-zA-Z0-9]*[\\._/a-zA-Z0-9\\-\\&\\?\\=]*$",
							},
						},
					},
					"sslConfiguration": {
						Properties: map[string]apiextv1beta1.JSONSchemaProps{
							"certificateName": {
								Type:    common.ValidationTypeString,
								Pattern: common.HostnameValidationRegex,
							},
							"verifyDepth": {
								Type:    common.ValidationTypeInteger,
								Minimum: &minDepth,
								// TODO: get proper max
								Maximum: &maxDepth,
							},
						},
					},
					"sessionPersistenceConfiguration": {
						Properties: map[string]apiextv1beta1.JSONSchemaProps{
							"cookieName": {
								Type:    common.ValidationTypeString,
								Pattern: common.HostnameValidationRegex,
							},
							"disableFallback": {
								Type: common.ValidationTypeBoolean,
							},
						},
					},
				},
			},
		},
	},
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BackendSet describes a backend set
type BackendSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              BackendSetSpec   `json:"spec"`
	Status            BackendSetStatus `json:"status,omitempty"`
}

// BackendSetSpec describes the backend set spec
type BackendSetSpec struct {
	LoadBalancerRef string `json:"loadBalancerRef"`

	HealthChecker *HealthChecker `json:"healthChecker"`

	Policy string `json:"policy" url:"-"` // FIXME: supposedly has default: "ROUND_ROBIN" but then raises error when null. For valid values see ListPolicies()

	SSLConfig                *SSLConfiguration                `json:"sslConfiguration,omitempty" url:"-"` // TODO: acc test, waiting on CreateCertificate() tests
	SessionPersistenceConfig *SessionPersistenceConfiguration `json:"sessionPersistenceConfiguration,omitempty" url:"-"`

	common.Dependency
}

// HealthChecker describes health checker of the backend set
type HealthChecker struct {
	Protocol string `url:"-" header:"-" json:"protocol"` // TODO: add validation in provider, must be in {"HTTP","TCP"}
	URLPath  string `url:"-" header:"-" json:"urlPath"`

	// Optional
	IntervalInMillis  int    `url:"-" header:"-" json:"intervalInMillis,omitempty"`  // Default: 10000
	Port              int    `url:"-" header:"-" json:"port,omitempty"`              // Default: 0
	ResponseBodyRegex string `url:"-" header:"-" json:"responseBodyRegex,omitempty"` // Default: ".*",
	Retries           int    `url:"-" header:"-" json:"retries,omitempty"`           // Default: 3
	ReturnCode        int    `url:"-" header:"-" json:"returnCode,omitempty"`        // Default: 200
	TimeoutInMillis   int    `url:"-" header:"-" json:"timeoutInMillis,omitempty"`   // Default: 3000,
}

// SSLConfiguration describes the ssl configuration of the backend set
type SSLConfiguration struct {
	CertificateName       string `json:"certificateName"`
	VerifyDepth           int    `json:"verifyDepth"`
	VerifyPeerCertificate bool   `json:"verifyPeerCertificate"`
}

// SessionPersistenceConfiguration descrbies the session persistence of the backend set
type SessionPersistenceConfiguration struct {
	CookieName      string `json:"cookieName"`
	DisableFallback bool   `json:"disableFallback"`
}

// BackendSetStatus describes the backend set status
type BackendSetStatus struct {
	common.ResourceStatus
	LoadBalancerId *string

	WorkRequestId     *string                              `json:"workRequestId,omitempty"`
	WorkRequestStatus *ocilb.WorkRequestLifecycleStateEnum `json:"workRequestStatus,omitempty"`

	Resource *BackendSetResource `json:"resource,omitempty"`
}

// BackendSetResource describes the backend set resource from oci
type BackendSetResource struct {
	*ocilb.BackendSet
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BackendSetList is a list of BackendSet items
type BackendSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []BackendSet `json:"items"`
}

// IsResource returns true if there is an oci id, otherwise false
func (s *BackendSet) IsResource() bool {
	if s.GetResourceID() == "" {
		return false
	}
	return true
}

// GetResourceID returns the oci id of the backend set
func (s *BackendSet) GetResourceID() string {
	var id string
	if s.Status.Resource != nil && s.Status.Resource.Name != nil {
		id = *s.Status.Resource.Name
	}
	return id
}

// GetResourcePlural returns the plural name of the backend set type
func (s *BackendSet) GetResourcePlural() string {
	return BackendSetResourcePlural
}

// GetGroupVersionResource returns the group name of the backend set type
func (s *BackendSet) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(BackendSetResourcePlural)
}

// SetResource sets the resource in the backend set status
func (s *BackendSet) SetResource(r *ocilb.BackendSet) *BackendSet {
	if r != nil {
		s.Status.Resource = &BackendSetResource{r}
	}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *BackendSet) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds a backend set dependent
func (s *BackendSet) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes a backend set dependent
func (s *BackendSet) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the backend set dependent is registered
func (s *BackendSet) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the backend set spec
func (in *BackendSetSpec) DeepCopy() *BackendSetSpec {
	if in == nil {
		return nil
	}
	out := in
	return out
}

// DeepCopy the backend set oci resource
func (in *BackendSetResource) DeepCopy() (out *BackendSetResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
