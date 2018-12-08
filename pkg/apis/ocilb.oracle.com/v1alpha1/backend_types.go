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

// Backend names
const (
	BackendKind           = "Backend"
	BackendResourcePlural = "backends"
	BackendControllerName = "backends"
)

var minTCPPort = float64(1)
var maxTCPPort = float64(65535)

var minWeight = float64(1)
var maxWeight = float64(100)

// BackendValidation describes the backend validation schema
var BackendValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"backendSetRef", "instanceRef", "loadBalancerRef", "port", "weight"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"backendSetRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"instanceRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"loadBalancerRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},

					"ipAddress": {
						Type:    common.ValidationTypeString,
						Pattern: common.Ipv4ValidationRegex,
					},
					"port": {
						Type:    common.ValidationTypeInteger,
						Minimum: &minTCPPort,
						Maximum: &maxTCPPort,
					},
					"weight": {
						Type:    common.ValidationTypeInteger,
						Minimum: &minWeight,
						Maximum: &maxWeight,
					},
					"backup": {
						Type: common.ValidationTypeBoolean,
					},
					"drain": {
						Type: common.ValidationTypeBoolean,
					},
					"online": {
						Type: common.ValidationTypeBoolean,
					},
				},
			},
		},
	},
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Backend describes a backend
type Backend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              BackendSpec   `json:"spec"`
	Status            BackendStatus `json:"status,omitempty"`
}

// BackendSpec describes a backend spec
type BackendSpec struct {
	BackendSetRef   string `json:"backendSetRef"`
	InstanceRef     string `json:"instanceRef"`
	LoadBalancerRef string `json:"loadBalancerRef"`

	Backup    bool   `json:"backup"`
	Drain     bool   `json:"drain"`
	IPAddress string `json:"ipAddress"`
	Offline   bool   `json:"offline"`
	Port      int    `json:"port"`
	Weight    int    `json:"weight"`

	common.Dependency
}

// BackendStatus describes a backend status
type BackendStatus struct {
	common.ResourceStatus
	LoadBalancerId *string

	WorkRequestId     *string                              `json:"workRequestId,omitempty"`
	WorkRequestStatus *ocilb.WorkRequestLifecycleStateEnum `json:"workRequestStatus,omitempty"`

	Resource *BackendResource `json:"resource,omitempty"`
}

// BackendResource describes a backend resource from oci
type BackendResource struct {
	*ocilb.Backend
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BackendList is a list of Backend items
type BackendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Backend `json:"items"`
}

// IsResource returns true if there is an oci id, otherwise false
func (s *Backend) IsResource() bool {
	if s.GetResourceID() == "" {
		return false
	}
	return true
}

// GetResourceID returns the oci id of the backend
func (s *Backend) GetResourceID() string {
	var id string
	if s.Status.Resource != nil && s.Status.Resource.Name != nil {
		id = *s.Status.Resource.Name
	}
	return id
}

// GetResourcePlural returns the plural name of the backend type
func (s *Backend) GetResourcePlural() string {
	return BackendResourcePlural
}

// GetGroupVersionResource returns the group version of the backend type
func (s *Backend) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(BackendResourcePlural)
}

// SetResource sets the resource in backend status
func (s *Backend) SetResource(r *ocilb.Backend) *Backend {
	if r != nil {
		s.Status.Resource = &BackendResource{r}
	}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *Backend) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds a backend dependent
func (s *Backend) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes a backend dependent
func (s *Backend) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the backend dependent is registered
func (s *Backend) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the backend spec
func (in *BackendSpec) DeepCopy() *BackendSpec {
	if in == nil {
		return nil
	}
	out := in
	return out
}

// DeepCopy the backed oci resource
func (in *BackendResource) DeepCopy() (out *BackendResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
