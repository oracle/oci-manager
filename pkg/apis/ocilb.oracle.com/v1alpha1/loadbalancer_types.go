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

// LoadBalancer names
const (
	LoadBalancerKind           = "LoadBalancer"
	LoadBalancerResourcePlural = "loadbalancers"
	LoadBalancerControllerName = "loadbalancers"
)

// LoadBalancerValidation describes the load balancer validation schema
var LoadBalancerValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"compartmentRef", "subnetRefs"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"compartmentRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"subnetRefs": {
						Type: common.ValidationTypeArray,
					},

					"isPrivate": {
						Type: common.ValidationTypeBoolean,
					},
					"shapeName": {
						Type:    common.ValidationTypeString,
						Pattern: "^100Mbps$|^400Mbps$|^8000Mbps$",
					},
				},
			},
		},
	},
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LoadBalancer describes a load balancer
type LoadBalancer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              LoadBalancerSpec   `json:"spec"`
	Status            LoadBalancerStatus `json:"status,omitempty"`
}

// LoadBalancerSpec describes a load balancer spec
type LoadBalancerSpec struct {
	CompartmentRef string   `json:"compartmentRef"`
	SubnetRefs     []string `json:"subnetRefs"`

	IsPrivate bool   `json:"isPrivate"`
	Shape     string `json:"shapeName"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	common.Dependency
}

// LoadBalancerStatus describes a load balancer status
type LoadBalancerStatus struct {
	common.ResourceStatus

	WorkRequestId     *string                              `json:"workRequestId,omitempty"`
	WorkRequestStatus *ocilb.WorkRequestLifecycleStateEnum `json:"workRequestStatus,omitempty"`

	Resource *LoadBalancerResource `json:"resource,omitempty"`
}

// LoadBalancerResource describes a load balancer resource from oci
type LoadBalancerResource struct {
	*ocilb.LoadBalancer
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LoadBalancerList is a list of LoadBalancer items
type LoadBalancerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []LoadBalancer `json:"items"`
}

// IsResource returns true if there is oci id and status is active, otherwise false
func (s *LoadBalancer) IsResource() bool {
	if s.GetResourceID() == "" {
		return false
	}
	return true
}

// GetResourceID returns the oci id of the load balancer
func (s *LoadBalancer) GetResourceID() string {
	var id string
	if s.Status.Resource != nil {
		id = *s.Status.Resource.Id
	}
	return id
}

// GetResourcePlural returns the plural name of the load balancer type
func (s *LoadBalancer) GetResourcePlural() string {
	return LoadBalancerResourcePlural
}

// GetGroupVersionResource returns the group version of the load balancer type
func (s *LoadBalancer) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(LoadBalancerResourcePlural)
}

// GetResourceLifecycleState returns the state of the load balancer
func (s *LoadBalancer) GetResourceLifecycleState() string {
	var state string
	if s.Status.Resource != nil {
		state = string(s.Status.Resource.LifecycleState)
	}
	return state
}

// SetResource set the resource in load balancer status
func (s *LoadBalancer) SetResource(r *ocilb.LoadBalancer) *LoadBalancer {
	if r != nil {
		s.Status.Resource = &LoadBalancerResource{r}
	}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *LoadBalancer) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds a load balancer dependent
func (s *LoadBalancer) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes a load balancer dependent
func (s *LoadBalancer) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the load balancer dependent is registered
func (s *LoadBalancer) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the load balancer spec
func (in *LoadBalancerSpec) DeepCopy() *LoadBalancerSpec {
	if in == nil {
		return nil
	}
	out := in
	return out
}

// DeepCopy the load balancer oci resource
func (in *LoadBalancerResource) DeepCopy() (out *LoadBalancerResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
