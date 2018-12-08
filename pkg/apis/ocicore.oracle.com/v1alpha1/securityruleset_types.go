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

// SecurityRuleSet names
// NOTE: The name SecurityRuleSet is different from the OCI resource name SecurityList
// because k8s does not support CRD names that end with "List"
const (
	SecurityRuleSetKind           = "SecurityRuleSet"
	SecurityRuleSetResourcePlural = "securityrulesets"
	SecurityRuleSetControllerName = "securityrulesets"
)

// SecurityRuleSetValidation describes the security rule set validation schema
var SecurityRuleSetValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"compartmentRef", "vcnRef", "egressSecurityRules", "ingressSecurityRules"},
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
					"egressSecurityRules": {
						Type: common.ValidationTypeArray,
					},
					"ingressSecurityRules": {
						Type: common.ValidationTypeArray,
					}},
			},
		},
	},
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SecurityRuleSet describes a security rule set
type SecurityRuleSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SecurityRuleSetSpec   `json:"spec"`
	Status            SecurityRuleSetStatus `json:"status,omitempty"`
}

// SecurityRuleSetSpec describes a security rule set spec
type SecurityRuleSetSpec struct {
	CompartmentRef string `json:"compartmentRef"`
	VcnRef         string `json:"vcnRef"`
	DisplayName    string `json:"displayName,omitempty"`

	EgressSecurityRules  []ocisdkcore.EgressSecurityRule  `json:"egressSecurityRules"`
	IngressSecurityRules []ocisdkcore.IngressSecurityRule `json:"ingressSecurityRules"`
	common.Dependency
}

// SecurityRuleSetStatus describes a security rule set status
type SecurityRuleSetStatus struct {
	common.ResourceStatus
	Resource *SecurityRuleSetResource `json:"resource,omitempty"`
}

// SecurityRuleSetResource describes a security rule set resource from oci
type SecurityRuleSetResource struct {
	ocisdkcore.SecurityList
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SecurityRuleSetList is a list of SecurityRuleSet items
type SecurityRuleSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SecurityRuleSet `json:"items"`
}

// IsResource returns true if there is an oci id, oterwise false
func (s *SecurityRuleSet) IsResource() bool {

	if s.GetResourceID() != "" && s.Status.Resource != nil && s.Status.Resource.LifecycleState == ocisdkcore.SecurityListLifecycleStateAvailable {
		return true
	}
	return false
}

// GetResourceID returns the oci id of the security rule set
func (s *SecurityRuleSet) GetResourceID() string {
	if s.Status.Resource != nil && s.Status.Resource.Id != nil {
		return *s.Status.Resource.Id
	}
	return ""
}

// GetResourcePlural returns the plural name of the security rule set type
func (s *SecurityRuleSet) GetResourcePlural() string {
	return SecurityRuleSetResourcePlural
}

// GetGroupVersionResource returns the group version name of the security rule set type
func (s *SecurityRuleSet) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(SecurityRuleSetResourcePlural)
}

// SetResource sets the resource in status of the security rule set
func (s *SecurityRuleSet) SetResource(r *ocisdkcore.SecurityList) *SecurityRuleSet {
	if r != nil {
		s.Status.Resource = &SecurityRuleSetResource{*r}
	}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *SecurityRuleSet) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds a security rule set dependent
func (s *SecurityRuleSet) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes a security rule set dependent
func (s *SecurityRuleSet) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the security rule set dependent is registered
func (s *SecurityRuleSet) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the security rule set spec
func (in *SecurityRuleSetSpec) DeepCopy() (out *SecurityRuleSetSpec) {
	if in == nil {
		return nil
	}
	out = in
	return
}

// DeepCopy the security rule set oci resource
func (in *SecurityRuleSetResource) DeepCopy() (out *SecurityRuleSetResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
