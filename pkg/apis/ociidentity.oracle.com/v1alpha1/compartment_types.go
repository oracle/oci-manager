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
	ocisdkidentity "github.com/oracle/oci-go-sdk/identity"
	common "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Compartment names
const (
	CompartmentKind           = "Compartment"
	CompartmentResourcePlural = "compartments"
	CompartmentControllerName = "compartments"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Compartment describes a compartment
type Compartment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              CompartmentSpec   `json:"spec"`
	Status            CompartmentStatus `json:"status,omitempty"`
}

// CompartmentSpec describes a compartment spec
type CompartmentSpec struct {
	Description string `json:"description,omitempty"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	common.Dependency
}

// CompartmentStatus describes a compartment status
type CompartmentStatus struct {
	common.ResourceStatus
	Shapes              []string             `json:"shapes,omitempty"`
	Images              map[string]string    `json:"images,omitempty"`
	AvailabilityDomains []string             `json:"availabilityDomains,omitempty"`
	Resource            *CompartmentResource `json:"resource,omitempty"`
}

// CompartmentResource describes a compartment resource from oci
type CompartmentResource struct {
	ocisdkidentity.Compartment
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//CompartmentList is a list of Compartment items
type CompartmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Compartment `json:"items"`
}

// IsResource returns true if there is oci id, otherwise false
func (s *Compartment) IsResource() bool {
	if s.GetResourceID() == "" {
		return false
	}
	return true
}

// GetResourceID returns the oci id of the compartment
func (s *Compartment) GetResourceID() string {
	if s.Status.Resource != nil && s.Status.Resource.Id != nil {
		return *s.Status.Resource.Id
	}
	return ""
}

// GetResourcePlural returns the plural name of the compartment type
func (s *Compartment) GetResourcePlural() string {
	return CompartmentResourcePlural
}

// GetGroupVersionResource returns the group version of the compartment type
func (s *Compartment) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(CompartmentResourcePlural)
}

// SetResource sets the resource in compartment status
func (s *Compartment) SetResource(r *ocisdkidentity.Compartment) *Compartment {
	if r != nil {
		s.Status.Resource = &CompartmentResource{*r}
	}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *Compartment) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds a compartment dependent
func (s *Compartment) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes a compartment dependent
func (s *Compartment) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the compartment dependent is registered
func (s *Compartment) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the load balancer spec
func (in *CompartmentSpec) DeepCopy() *CompartmentSpec {
	if in == nil {
		return nil
	}
	out := in
	return out
}

// DeepCopy the compartment oci resource
func (in *CompartmentResource) DeepCopy() (out *CompartmentResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
