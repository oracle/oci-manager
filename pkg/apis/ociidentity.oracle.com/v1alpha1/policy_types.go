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
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	ocisdkidentity "github.com/oracle/oci-go-sdk/identity"
	common "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Policy names
const (
	PolicyKind           = "Policy"
	PolicyResourcePlural = "policies"
	PolicyControllerName = "policies"
)

// PolicyValidation describes the policy validation schema
var PolicyValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"compartmentRef", "description", "statements"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"compartmentRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"description": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"statements": {
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

// Policy describes a policy
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              PolicySpec   `json:"spec"`
	Status            PolicyStatus `json:"status,omitempty"`
}

// PolicySpec describes a policy spec
type PolicySpec struct {
	CompartmentRef string `json:"compartmentRef"`

	// The description you assign to the policy. Does not have to be unique, and it's changeable.
	Description *string `mandatory:"true" json:"description"`

	// An array of one or more policy statements written in the policy language.
	Statements []string `mandatory:"true" json:"statements"`

	common.Dependency
}

// PolicyStatus describes a policy status
type PolicyStatus struct {
	common.ResourceStatus
	Resource *PolicyResource `json:"resource,omitempty"`
}

// PolicyResource describes a policy resource from oci
type PolicyResource struct {
	ocisdkidentity.Policy
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolicyList is a list of Policy resources
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Policy `json:"items"`
}

// IsResource returns true if there is an oci id and state is available, otherwise false
func (s *Policy) IsResource() bool {
	if s.GetResourceID() == "" {
		return false
	}
	return true
}

// GetResourceID returns oci id of the policy
func (s *Policy) GetResourceID() string {
	if s.Status.Resource != nil && s.Status.Resource.Id != nil {
		return *s.Status.Resource.Id
	}
	return ""
}

// GetResourcePlural returns plural name of the policy type
func (s *Policy) GetResourcePlural() string {
	return PolicyResourcePlural
}

// GetGroupVersionResource returns group version of the policy type
func (s *Policy) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(PolicyResourcePlural)
}

// SetResource sets the resource in the status of the policy
func (s *Policy) SetResource(r *ocisdkidentity.Policy) *Policy {
	if r != nil {
		s.Status.Resource = &PolicyResource{*r}
	}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *Policy) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds a policy dependent
func (s *Policy) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes a policy dependent
func (s *Policy) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the policy dependent is registered
func (s *Policy) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the policy oci resource
func (in *PolicyResource) DeepCopy() (out *PolicyResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
