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

// DhcpOption names
const (
	DhcpOptionKind           = "DhcpOption"
	DhcpOptionResourcePlural = "dhcpoptions"
	DhcpOptionControllerName = "dhcpoptions"
)

// DhcpOptionValidation describes the dhcp options validation schema
var DhcpOptionValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"compartmentRef", "vcnRef"},
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
					"options": {
						Type: common.ValidationTypeArray,
					},
				},
			},
		},
	},
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DhcpOption describes a dhcp options
type DhcpOption struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   DhcpOptionSpec   `json:"spec"`
	Status DhcpOptionStatus `json:"status,omitempty"`
}

// DhcpOptionSpec describes a dhcp options spec
type DhcpOptionSpec struct {
	CompartmentRef string `json:"compartmentRef"`
	VcnRef         string `json:"vcnRef"`

	DisplayName string                  `json:"displayName,omitempty"`
	Options     []ocisdkcore.DhcpOption `json:"options"`

	common.Dependency
}

// DhcpOptionStatus describes a dhcp options status
type DhcpOptionStatus struct {
	common.ResourceStatus
	Resource *DhcpOptionResource `json:"resource,omitempty"`
}

// DhcpOptionResource describes a dhcp options resource from oci
type DhcpOptionResource struct {
	ocisdkcore.DhcpOptions
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DhcpOptionList is a list of DhcpOption items
type DhcpOptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []DhcpOption `json:"items"`
}

// IsResource returns true if there is an oci id and state is available, otherwise false
func (s *DhcpOption) IsResource() bool {
	if s.GetResourceID() == "" {
		return false
	}
	return true
}

// GetResourceID returns oci id of the dhcp options
func (s *DhcpOption) GetResourceID() string {
	if s.Status.Resource != nil && s.Status.Resource.Id != nil {
		return *s.Status.Resource.Id
	}
	return ""
}

// GetResourcePlural returns plural name of the dhcp options type
func (s *DhcpOption) GetResourcePlural() string {
	return DhcpOptionResourcePlural
}

// GetGroupVersionResource returns group version of the dhcp options type
func (s *DhcpOption) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(DhcpOptionResourcePlural)
}

// AddDependent adds a dhcp options dependent
func (s *DhcpOption) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes a dhcp options dependent
func (s *DhcpOption) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the dhcp options dependent is registered
func (s *DhcpOption) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// SetResource sets the resource in the status of the dhcp options
func (s *DhcpOption) SetResource(r *ocisdkcore.DhcpOptions) *DhcpOption {
	if r != nil {
		s.Status.Resource = &DhcpOptionResource{*r}
	}
	return s
}

// GetResourceState returns the current state of the resource
func (s *DhcpOption) GetResourceState() common.ResourceState {
	return s.Status.State
}

// DeepCopy the dhcp options spec
func (in *DhcpOptionSpec) DeepCopy() (out *DhcpOptionSpec) {
	if in == nil {
		return nil
	}
	out = in
	return
}

// DeepCopy the dhcp options oci resource
func (in *DhcpOptionResource) DeepCopy() (out *DhcpOptionResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
