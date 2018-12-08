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

// RouteTable names
const (
	RouteTableKind           = "RouteTable"
	RouteTableResourcePlural = "routetables"
	RouteTableControllerName = "routetables"
)

// RouteTableValidation describes the route table validation schema
var RouteTableValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"compartmentRef", "vcnRef", "routeRules"},
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
					"routeRules": {
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

// RouteTable describes a route table
type RouteTable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              RouteTableSpec   `json:"spec"`
	Status            RouteTableStatus `json:"status,omitempty"`
}

// RouteTableSpec describes a route table spec
type RouteTableSpec struct {
	CompartmentRef string      `json:"compartmentRef"`
	VcnRef         string      `json:"vcnRef"`
	DisplayName    string      `json:"displayName,omitempty"`
	RouteRules     []RouteRule `json:"routeRules"`
	common.Dependency
}

// RouteTableStatus describes a route table status
type RouteTableStatus struct {
	common.ResourceStatus
	Resource *RouteTableResource `json:"resource,omitempty"`
}

// RouteTableResource describes a route table resource from oci
type RouteTableResource struct {
	ocisdkcore.RouteTable
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RouteTableList is a list of RouteTable items
type RouteTableList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []RouteTable `json:"items"`
}

// RouteRule describes a route rule in the route table
type RouteRule struct {
	CidrBlock       string `json:"cidrBlock"`
	NetworkEntityID string `json:"networkEntityId"`
}

// IsResource returns true if there is an oci id, otherwise false
func (s *RouteTable) IsResource() bool {

	if s.GetResourceID() != "" && s.Status.Resource != nil && s.Status.Resource.LifecycleState == ocisdkcore.RouteTableLifecycleStateAvailable {
		return true
	}
	return false
}

// GetResourceID returns the oci id of the route table
func (s *RouteTable) GetResourceID() string {
	if s.Status.Resource != nil && s.Status.Resource.Id != nil {
		return *s.Status.Resource.Id
	}
	return ""
}

// GetResourcePlural returns the plural name of the route table type
func (s *RouteTable) GetResourcePlural() string {
	return RouteTableResourcePlural
}

// GetGroupVersionResource returns the group version of the route table type
func (s *RouteTable) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(RouteTableResourcePlural)
}

// SetResource sets the resource in the status of the route table
func (s *RouteTable) SetResource(r *ocisdkcore.RouteTable) *RouteTable {
	if r != nil {
		s.Status.Resource = &RouteTableResource{*r}
	}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *RouteTable) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds a route table dependent
func (s *RouteTable) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes a route table dependent
func (s *RouteTable) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the route table dependent is registered
func (s *RouteTable) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the route table oci resource
func (in *RouteTableResource) DeepCopy() (out *RouteTableResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
