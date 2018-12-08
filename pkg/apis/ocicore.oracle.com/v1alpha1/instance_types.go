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

// Instance names
const (
	InstanceKind           = "Instance"
	InstanceResourcePlural = "instances"
	InstanceControllerName = "instances"
)

// InstanceValidation describes the instance validation schema
var InstanceValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"compartmentRef", "subnetRef", "image", "shape"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"compartmentRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"subnetRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"availabilityDomain": {
						Type:    common.ValidationTypeString,
						Pattern: common.AvailabilityDomainValidationRegex,
					},
					"displayName": {
						Type:    common.ValidationTypeString,
						Pattern: common.HostnameValidationRegex,
					},
					"hostnameLabel": {
						Type:    common.ValidationTypeString,
						Pattern: common.HostnameValidationRegex,
					},
					"image": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"ipxeScript": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"shape": {
						Type:    common.ValidationTypeString,
						Pattern: "^BM\\.|^VM\\.",
					},
				},
			},
		},
	},
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Instance describes an instance
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              InstanceSpec   `json:"spec"`
	Status            InstanceStatus `json:"status,omitempty"`
}

// InstanceSpec describes an instance spec
type InstanceSpec struct {
	CompartmentRef string `json:"compartmentRef"`
	SubnetRef      string `json:"subnetRef"`

	AvailabilityDomain string `json:"availabilityDomain"`
	DisplayName        string `json:"displayName,omitempty"`
	HostnameLabel      string `json:"hostnameLabel,omitempty"`
	Image              string `json:"image"`
	IpxeScript         string `json:"ipxeScript,omitempty"`
	Shape              string `json:"shape"`

	Metadata         map[string]string      `json:"metadata,omitempty"`
	ExtendedMetadata map[string]interface{} `json:"extendedMetadata,omitempty"`
	common.Dependency
}

// InstanceStatus describes an instance status
type InstanceStatus struct {
	common.ResourceStatus
	Resource    *InstanceResource    `json:"resource,omitempty"`
	PrimaryVnic *PrimaryVnicResource `json:"primaryVnic,omitempty"`
	BootVolume  *BootVolumeResource  `json:"bootVolume,omitempty"`
}

// InstanceResource describes an instance resource from oci
type InstanceResource struct {
	ocisdkcore.Instance
}

// PrimaryVnicResource describes the primary vnic associated with the instance from oci
type PrimaryVnicResource struct {
	ocisdkcore.Vnic
}

// BootVolumeResource describes the boot volume associated with the instance from oci
type BootVolumeResource struct {
	ocisdkcore.BootVolume
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InstanceList is a list of Instance items
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Instance `json:"items"`
}

// IsResource returns true if there is an oci id and it's in a running state, otherwise false
func (s *Instance) IsResource() bool {
	if s.GetResourceID() != "" && s.GetResourceLifecycleState() == string(ocisdkcore.InstanceLifecycleStateRunning) {
		return true
	}
	return false
}

// GetResourceID returns the oci id of the instance
func (s *Instance) GetResourceID() string {
	if s.Status.Resource != nil && s.Status.Resource.Id != nil {
		return *s.Status.Resource.Id
	}
	return ""
}

// GetResourcePlural returns the plural name of instance type
func (s *Instance) GetResourcePlural() string {
	return InstanceResourcePlural
}

// GetGroupVersionResource returns the group version of the instance type
func (s *Instance) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(InstanceResourcePlural)
}

// GetResourceLifecycleState returns the current state of the instance
func (s *Instance) GetResourceLifecycleState() string {
	var state string
	if s.Status.Resource != nil {
		state = string(s.Status.Resource.LifecycleState)
	}
	return state
}

// GetResourceState returns the current state of the iresource
func (s *Instance) GetResourceState() common.ResourceState {
	return s.Status.State
}

// SetResource sets the resource in status of the instance
func (s *Instance) SetResource(r *ocisdkcore.Instance) *Instance {
	if r != nil {
		s.Status.Resource = &InstanceResource{*r}
	}
	return s
}

// AddDependent adds an instance dependent
func (s *Instance) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes an instance dependent
func (s *Instance) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the instance dependent is registered
func (s *Instance) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the instance spec
func (in *InstanceSpec) DeepCopy() *InstanceSpec {
	if in == nil {
		return nil
	}
	out := in
	return out
}

// DeepCopy the instance oci resource
func (in *InstanceResource) DeepCopy() (out *InstanceResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}

// DeepCopy the primary vnic oci resource of the instance
func (in *PrimaryVnicResource) DeepCopy() (out *PrimaryVnicResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}

// DeepCopy the boot volume oci resource of the instance
func (in *BootVolumeResource) DeepCopy() (out *BootVolumeResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
