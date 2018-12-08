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

// VolumeBackup names
const (
	VolumeBackupKind           = "VolumeBackup"
	VolumeBackupResourcePlural = "volumebackups"
	VolumeBackupControllerName = "volumebackups"
)

var minVolumeBackupSizeInGBs = float64(50)
var maxVolumeBackupSizeInGBs = float64(16000)

// VolumeBackupValidation describes the volume backup validation schema
var VolumeBackupValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"volumeRef", "type"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"displayName": {
						Type:    common.ValidationTypeString,
						Pattern: common.HostnameValidationRegex,
					},
					"type": {
						Type:    common.ValidationTypeString,
						Pattern: "^FULL$|^INCREMENTAL$",
					},
					"volumeRef": {
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

// VolumeBackup describes a volume backup
type VolumeBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              VolumeBackupSpec   `json:"spec"`
	Status            VolumeBackupStatus `json:"status,omitempty"`
}

// VolumeBackupSpec describes a volume backup spec
type VolumeBackupSpec struct {
	VolumeRef string `json:"volumeRef,omitempty"`

	DisplayName      string `json:"displayName,omitempty"`
	VolumeBackupType string `json:"type"`

	common.Dependency
}

// VolumeBackupStatus describes a volume backup status
type VolumeBackupStatus struct {
	common.ResourceStatus
	Resource *VolumeBackupResource `json:"resource,omitempty"`
}

// VolumeBackupResource describes a volume backup resource from oci
type VolumeBackupResource struct {
	ocisdkcore.VolumeBackup
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeBackupList is a list of VolumeBackup items
type VolumeBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []VolumeBackup `json:"items"`
}

// IsResource returns true if there is an oci id, otherwise false
func (s *VolumeBackup) IsResource() bool {
	if s.GetResourceID() != "" && s.GetResourceLifecycleState() == string(ocisdkcore.VolumeBackupLifecycleStateAvailable) {
		return true
	}
	return false
}

// GetResourceID returns the oci id of the volume backup
func (s *VolumeBackup) GetResourceID() string {
	if s.Status.Resource != nil && s.Status.Resource.Id != nil {
		return *s.Status.Resource.Id
	}
	return ""
}

// GetResourcePlural returns the plural name of the volume backup type
func (s *VolumeBackup) GetResourcePlural() string {
	return VolumeBackupResourcePlural
}

// GetGroupVersionResource returns the group version of the volume backup type
func (s *VolumeBackup) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(VolumeBackupResourcePlural)
}

// GetResourceLifecycleState returns the volume backup state
func (s *VolumeBackup) GetResourceLifecycleState() string {
	var state string
	if s.Status.Resource != nil {
		state = string(s.Status.Resource.LifecycleState)
	}
	return state
}

// SetResource sets the resource in the status of the volume backup
func (s *VolumeBackup) SetResource(r *ocisdkcore.VolumeBackup) *VolumeBackup {
	if r != nil {
		s.Status.Resource = &VolumeBackupResource{*r}
	}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *VolumeBackup) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds a volume backup dependent
func (s *VolumeBackup) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes a volume backup dependent
func (s *VolumeBackup) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the volume backup dependent is registered
func (s *VolumeBackup) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the volume backup spec
func (in *VolumeBackupSpec) DeepCopy() *VolumeBackupSpec {
	if in == nil {
		return nil
	}
	out := in
	return out
}

// DeepCopy the volume backup oci resource
func (in *VolumeBackupResource) DeepCopy() (out *VolumeBackupResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
