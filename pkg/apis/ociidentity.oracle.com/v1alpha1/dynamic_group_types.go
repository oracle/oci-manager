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

// DynamicGroup names
const (
	DynamicGroupKind           = "DynamicGroup"
	DynamicGroupResourcePlural = "dynamicgroups"
	DynamicGroupControllerName = "dynamicgroups"
)

// DynamicGroupValidation describes the dynamic group validation schema
var DynamicGroupValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"compartmentRef", "description", "matchingRule"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"compartmentRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"description": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"matchingRule": {
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

// DynamicGroup describes a dynamic group
type DynamicGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              DynamicGroupSpec   `json:"spec"`
	Status            DynamicGroupStatus `json:"status,omitempty"`
}

// DynamicGroupSpec describes a dynamic group spec
type DynamicGroupSpec struct {
	CompartmentRef string `json:"compartmentRef"`

	// The description you assign to the group. Does not have to be unique, and it's changeable.
	Description *string `mandatory:"true" json:"description"`

	// A rule string that defines which instance certificates will be matched.
	// For syntax, see Managing Dynamic Groups (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Tasks/managingdynamicgroups.htm).
	MatchingRule *string `mandatory:"true" json:"matchingRule"`

	common.Dependency
}

// DynamicGroupStatus describes a dynamic group status
type DynamicGroupStatus struct {
	common.ResourceStatus
	Resource *DynamicGroupResource `json:"resource,omitempty"`
}

// DynamicGroupResource describes a dynamic group resource from oci
type DynamicGroupResource struct {
	ocisdkidentity.DynamicGroup
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DynamicGroupList is a list of DynamicGroup resources
type DynamicGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []DynamicGroup `json:"items"`
}

// IsResource returns true if there is an oci id and state is available, otherwise false
func (s *DynamicGroup) IsResource() bool {
	if s.GetResourceID() == "" {
		return false
	}
	return true
}

// GetResourceID returns oci id of the dynamic group
func (s *DynamicGroup) GetResourceID() string {
	if s.Status.Resource != nil && s.Status.Resource.Id != nil {
		return *s.Status.Resource.Id
	}
	return ""
}

// GetResourcePlural returns plural name of the dynamic group type
func (s *DynamicGroup) GetResourcePlural() string {
	return DynamicGroupResourcePlural
}

// GetGroupVersionResource returns group version of the dynamic group type
func (s *DynamicGroup) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(DynamicGroupResourcePlural)
}

// SetResource sets the resource in the status of the dynamic group
func (s *DynamicGroup) SetResource(r *ocisdkidentity.DynamicGroup) *DynamicGroup {
	if r != nil {
		s.Status.Resource = &DynamicGroupResource{*r}
	}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *DynamicGroup) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds a dynamic group dependent
func (s *DynamicGroup) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes a dynamic group dependent
func (s *DynamicGroup) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the dynamic group dependent is registered
func (s *DynamicGroup) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the dynamic group oci resource
func (in *DynamicGroupResource) DeepCopy() (out *DynamicGroupResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
