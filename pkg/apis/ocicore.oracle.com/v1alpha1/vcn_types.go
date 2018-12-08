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

// Vcn names
const (
	VirtualNetworkKind           = "Vcn"
	VirtualNetworkResourcePlural = "vcns"
	VirtualNetworkControllerName = "vcns"
)

// VcnValidation describes the vcn validation schema
var VcnValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"compartmentRef", "cidrBlock", "dnsLabel"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"compartmentRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"cidrBlock": {
						Type:    common.ValidationTypeString,
						Pattern: common.CidrValidationRegex,
					},
					"displayName": {
						Type:    common.ValidationTypeString,
						Pattern: common.HostnameValidationRegex,
					},
					"dnsLabel": {
						Type:    common.ValidationTypeString,
						Pattern: common.HostnameValidationRegex,
					},
					// can be empty string
					"vcnDomainName": {
						Type: common.ValidationTypeString,
					},
				},
			},
		},
	},
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//Vcn describes a vcn
type Vcn struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              VcnSpec   `json:"spec"`
	Status            VcnStatus `json:"status,omitempty"`
}

// VcnSpec describes a vcn spec
type VcnSpec struct {
	CompartmentRef string `json:"compartmentRef"`
	CidrBlock      string `json:"cidrBlock"`
	DisplayName    string `json:"displayName,omitempty"`
	DNSLabel       string `json:"dnsLabel"`
	VcnDomainName  string `json:"vcnDomainName"`
	common.Dependency
}

// VcnStatus describes a vcn status
type VcnStatus struct {
	common.ResourceStatus
	Resource *VcnResource `json:"resource,omitempty"`
}

// VcnResource describes a vcn resource from oci
type VcnResource struct {
	ocisdkcore.Vcn
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VcnList is a list of Vcn items
type VcnList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Vcn `json:"items"`
}

// IsResource returns true if there is an oci id and status is available, otherwise false
func (s *Vcn) IsResource() bool {
	if s.GetResourceID() != "" && s.GetResourceLifecycleState() == string(ocisdkcore.VcnLifecycleStateAvailable) {
		return true
	}
	return false
}

// GetResourceID returns the oci id of the vcn
func (s *Vcn) GetResourceID() string {
	if s.Status.Resource != nil && s.Status.Resource.Id != nil {
		return *s.Status.Resource.Id
	}
	return ""
}

// GetResourcePlural returns the plural name of the vcn type
func (s *Vcn) GetResourcePlural() string {
	return VirtualNetworkResourcePlural
}

// GetGroupVersionResource returns the group version of the vcn type
func (s *Vcn) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(VirtualNetworkResourcePlural)
}

// GetResourceLifecycleState returns the state of the vcn
func (s *Vcn) GetResourceLifecycleState() string {
	var state string
	if s.Status.Resource != nil {
		state = string(s.Status.Resource.LifecycleState)
	}
	return state
}

// SetResource sets the resource in vcn status
func (s *Vcn) SetResource(r *ocisdkcore.Vcn) *Vcn {
	if r != nil {
		s.Status.Resource = &VcnResource{*r}
	}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *Vcn) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds a vcn dependent
func (s *Vcn) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent remvoes a vcn dependent
func (s *Vcn) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the vcn dependent is registered
func (s *Vcn) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the vcn oci resource
func (in *VcnResource) DeepCopy() (out *VcnResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
