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
	ocisdkcore "github.com/oracle/oci-go-sdk/core"
	common "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// InternetGateway names
const (
	InternetGatewayKind           = "InternetGateway"
	InternetGatewayResourcePlural = "internetgatewaies"
	InternetGatewayControllerName = "internetgatewaies"
)

// InternetGatewayValidation describes the internet gateway validation schema
var InternetGatewayValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"compartmentRef", "vcnRef", "isEnabled"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"compartmentRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"vcnRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"displayName": {
						Type:    common.ValidationTypeString,
						Pattern: common.HostnameValidationRegex,
					},
					"isEnabled": {
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

// InternetGateway describes an internet gateway
type InternetGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              InternetGatewaySpec   `json:"spec"`
	Status            InternetGatewayStatus `json:"status,omitempty"`
}

// InternetGatewaySpec describes an internet gateway spec
type InternetGatewaySpec struct {
	CompartmentRef string `json:"compartmentRef"`
	VcnRef         string `json:"vcnRef"`
	DisplayName    string `json:"displayName,omitempty"`
	IsEnabled      bool   `json:"isEnabled"`
	common.Dependency
}

// InternetGatewayStatus describes an internet gateway status
type InternetGatewayStatus struct {
	common.ResourceStatus
	Resource *InternetGatewayResource `json:"resource,omitempty"`
}

// InternetGatewayResource describes an internet gateway resource from oci
type InternetGatewayResource struct {
	ocisdkcore.InternetGateway
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//InternetGatewayList is a list of InternetGateway items
type InternetGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []InternetGateway `json:"items"`
}

// IsResource returns true if there is an oci id, otherwise false
func (s *InternetGateway) IsResource() bool {
	if s.GetResourceID() != "" && s.Status.Resource != nil && s.Status.Resource.LifecycleState == ocisdkcore.InternetGatewayLifecycleStateAvailable {
		return true
	}
	return false
}

// GetResourceID returns the oci id of the internet gateway
func (s *InternetGateway) GetResourceID() string {
	if s.Status.Resource != nil && s.Status.Resource.Id != nil {
		return *s.Status.Resource.Id
	}
	return ""
}

// GetResourcePlural returns the plural name of internet gateway type
func (s *InternetGateway) GetResourcePlural() string {
	return InternetGatewayResourcePlural
}

// GetGroupVersionResource returns the group version of the internet gateway type
func (s *InternetGateway) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(InternetGatewayResourcePlural)
}

// SetResource sets the resource in status of the internet gateway
func (s *InternetGateway) SetResource(r *ocisdkcore.InternetGateway) *InternetGateway {
	if r != nil {
		s.Status.Resource = &InternetGatewayResource{*r}
	}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *InternetGateway) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds an internet gateway dependent
func (s *InternetGateway) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes an internet gateway dependent
func (s *InternetGateway) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the internet gateway dependent is registered
func (s *InternetGateway) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the internet gateway oci resource
func (in *InternetGatewayResource) DeepCopy() (out *InternetGatewayResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
