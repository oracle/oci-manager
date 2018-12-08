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

// Subnet names
const (
	SubnetKind           = "Subnet"
	SubnetResourcePlural = "subnets"
	SubnetControllerName = "subnets"
)

// SubnetValidation describes the subnet validation schema
var SubnetValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"compartmentRef", "routetableRef", "securityrulesetRefs", "vcnRef", "dnsLabel", "cidrBlock"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"compartmentRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"routetableRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"securityrulesetRefs": {
						Type: common.ValidationTypeArray,
					},
					"vcnRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"availabilityDomain": {
						Type:    common.ValidationTypeString,
						Pattern: common.AvailabilityDomainValidationRegex,
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
				},
			},
		},
	},
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Subnet describes a subnet
type Subnet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SubnetSpec   `json:"spec"`
	Status            SubnetStatus `json:"status,omitempty"`
}

// SubnetSpec describes a subnet spec
type SubnetSpec struct {
	CompartmentRef string `json:"compartmentRef"`
	VcnRef         string `json:"vcnRef"`

	AvailabilityDomain  string   `json:"availabilityDomain"`
	CidrBlock           string   `json:"cidrBlock"`
	DisplayName         string   `json:"displayName,omitempty"`
	DNSLabel            string   `json:"dnsLabel,omitempty"`
	RouteTableRef       string   `json:"routetableRef,omitempty"`
	SecurityRuleSetRefs []string `json:"securityrulesetRefs,omitempty"`
	common.Dependency
}

// SubnetStatus describes a subnet status
type SubnetStatus struct {
	common.ResourceStatus
	Resource *SubnetResource `json:"resource,omitempty"`
}

// SubnetResource describes a subnet resource from oci
type SubnetResource struct {
	ocisdkcore.Subnet
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SubnetList is a list of Subnet items
type SubnetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Subnet `json:"items"`
}

// IsResource returns true if there is an oci id, otherwise false
func (s *Subnet) IsResource() bool {
	if s.GetResourceID() != "" && s.GetResourceLifecycleState() == string(ocisdkcore.SubnetLifecycleStateAvailable) {
		return true
	}
	return false
}

// GetResourceID returns the oci id of the subnet
func (s *Subnet) GetResourceID() string {
	if s.Status.Resource != nil && s.Status.Resource.Id != nil {
		return *s.Status.Resource.Id
	}
	return ""
}

// GetResourcePlural returns the plural name of the subnet type
func (s *Subnet) GetResourcePlural() string {
	return SubnetResourcePlural
}

// GetGroupVersionResource returns the group version of the subnet type
func (s *Subnet) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(SubnetResourcePlural)
}

// GetResourceLifecycleState returns the state of the subnet
func (s *Subnet) GetResourceLifecycleState() string {
	var state string
	if s.Status.Resource != nil {
		state = string(s.Status.Resource.LifecycleState)
	}
	return state
}

// SetResource sets the resource in status of the subnet
func (s *Subnet) SetResource(r *ocisdkcore.Subnet) *Subnet {
	if r != nil {
		s.Status.Resource = &SubnetResource{*r}
	}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *Subnet) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds a subnet dependent
func (s *Subnet) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes a subnet dependent
func (s *Subnet) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the subnet dependent is registered
func (s *Subnet) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the subnet oci resource
func (in *SubnetResource) DeepCopy() (out *SubnetResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
