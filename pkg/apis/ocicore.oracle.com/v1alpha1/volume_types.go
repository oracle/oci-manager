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
	"encoding/json"

	"github.com/golang/glog"
	ocisdkcore "github.com/oracle/oci-go-sdk/core"
	common "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Volume names
const (
	VolumeKind           = "Volume"
	VolumeResourcePlural = "volumes"
	VolumeControllerName = "volumes"
)

var minVolumeSizeInGBs = float64(50)
var maxVolumeSizeInGBs = float64(16000)

// VolumeValidation describes the volume validation schema
var VolumeValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"compartmentRef", "instanceRef", "availabilityDomain", "sizeInGBs"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"compartmentRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"instanceRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"attachmentType": {
						Type:    common.ValidationTypeString,
						Pattern: "iscsi|paravirtualized",
					},
					"availabilityDomain": {
						Type:    common.ValidationTypeString,
						Pattern: common.AvailabilityDomainValidationRegex,
					},
					"displayName": {
						Type:    common.ValidationTypeString,
						Pattern: common.HostnameValidationRegex,
					},
					"sizeInGBs": {
						Type:    common.ValidationTypeInteger,
						Minimum: &minVolumeSizeInGBs,
						Maximum: &maxVolumeSizeInGBs,
					},
				},
			},
		},
	},
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Volume describes a volume
type Volume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              VolumeSpec   `json:"spec"`
	Status            VolumeStatus `json:"status,omitempty"`
}

// VolumeSpec describes a volume spec
type VolumeSpec struct {
	CompartmentRef string `json:"compartmentRef"`
	InstanceRef    string `json:"instanceRef"`

	DisplayName        string `json:"displayName,omitempty"`
	AvailabilityDomain string `json:"availabilityDomain"`
	SizeInGBs          int64  `json:"sizeInGBs"`
	AttachmentType     string `json:"attachmentType,omitempty"`
	common.Dependency
}

// VolumeStatus describes a volume status
type VolumeStatus struct {
	common.ResourceStatus
	Resource        *VolumeResource                               `json:"resource,omitempty"`
	AttachmentState ocisdkcore.VolumeAttachmentLifecycleStateEnum `json:"attachmentState,omitempty"`
	Attachment      *VolumeAttachment                             `json:"attachment,omitempty"`
}

// VolumeResource describes a volume resource from oci
type VolumeResource struct {
	ocisdkcore.Volume
}

// VolumeAttachment describes a volume attachment
type VolumeAttachment struct {
	AttachmentType string
	ocisdkcore.VolumeAttachment
}

// PVVolumeAttachment describes a paravirtualized volume attachment
type PVVolumeAttachment struct {
	VolumeAttachment ocisdkcore.ParavirtualizedVolumeAttachment
}

// SCSIVolumeAttachment describes an iscsi volume attachment
type SCSIVolumeAttachment struct {
	VolumeAttachment ocisdkcore.IScsiVolumeAttachment
}

// UnmarshalJSON is used to process the volume attachment type directly from json
// NOTE: there is an issue with processing objects as interfaces
func (m *VolumeAttachment) UnmarshalJSON(data []byte) error {

	type modeltogettype struct {
		AttachmentType string
	}
	model := modeltogettype{}
	err := json.Unmarshal(data, &model)
	if err != nil {
		glog.Errorf("Error unmarshal %s", err)
	}
	switch model.AttachmentType {
	case "iscsi":
		mm := SCSIVolumeAttachment{}
		err := json.Unmarshal(data, &mm)
		m.VolumeAttachment = mm.VolumeAttachment
		m.AttachmentType = model.AttachmentType
		return err
	case "paravirtualized":
		mm := PVVolumeAttachment{}
		err := json.Unmarshal(data, &mm)
		m.VolumeAttachment = mm.VolumeAttachment
		m.AttachmentType = model.AttachmentType
		return err
	default:
		mm := PVVolumeAttachment{}
		err := json.Unmarshal(data, &mm)
		m.VolumeAttachment = mm.VolumeAttachment
		m.AttachmentType = model.AttachmentType
		return err
	}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeList is a list of Volume items
type VolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Volume `json:"items"`
}

// IsResource returns true if there is an oci id and state is available, otherwise false
func (s *Volume) IsResource() bool {
	if s.GetResourceID() != "" && s.GetResourceLifecycleState() == string(ocisdkcore.VolumeLifecycleStateAvailable) {
		return true
	}
	return false
}

// GetResourceID returns oci id of the volume
func (s *Volume) GetResourceID() string {
	if s.Status.Resource != nil && s.Status.Resource.Id != nil {
		return *s.Status.Resource.Id
	}
	return ""
}

// GetResourcePlural returns plural name of the volume type
func (s *Volume) GetResourcePlural() string {
	return VolumeResourcePlural
}

// GetGroupVersionResource returns group version of the volume type
func (s *Volume) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(VolumeResourcePlural)
}

// GetResourceLifecycleState returns the volume state
func (s *Volume) GetResourceLifecycleState() string {
	var state string
	if s.Status.Resource != nil {
		state = string(s.Status.Resource.LifecycleState)
	}
	return state
}

// SetResource sets the resource in the status of the volume
func (s *Volume) SetResource(r *ocisdkcore.Volume) *Volume {
	if r != nil {
		s.Status.Resource = &VolumeResource{*r}
	}
	return s
}

// SetAttachment sets the volume attachment
func (s *Volume) SetAttachment(atype string, r *ocisdkcore.VolumeAttachment) *Volume {
	if r != nil {
		s.Status.Attachment = &VolumeAttachment{atype, *r}
		return s
	}
	s.Status.Attachment = &VolumeAttachment{}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *Volume) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds a volume dependent
func (s *Volume) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes a volume dependent
func (s *Volume) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the volume dependent is registered
func (s *Volume) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the volume spec
func (in *VolumeSpec) DeepCopy() *VolumeSpec {
	if in == nil {
		return nil
	}
	out := in
	return out
}

// DeepCopy the volume oci resource
func (in *VolumeResource) DeepCopy() (out *VolumeResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}

// DeepCopy the volume attachment
func (in *VolumeAttachment) DeepCopy() (out *VolumeAttachment) {
	if in == nil {
		return nil
	}
	out = in
	return
}

// DeepCopyInto the paravirtualized volume attachment
func (in *PVVolumeAttachment) DeepCopyInto(out *PVVolumeAttachment) {
	if in == nil {
		return
	}
	out = in
	return
}

// DeepCopy the paravirtualized volume attachment
func (in *PVVolumeAttachment) DeepCopy() *PVVolumeAttachment {
	if in == nil {
		return nil
	}
	out := new(PVVolumeAttachment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto the iscsi volume attachment
func (in *SCSIVolumeAttachment) DeepCopyInto(out *SCSIVolumeAttachment) {
	if in == nil {
		return
	}
	out = in
	return
}

// DeepCopy the iscsi volume attachment
func (in *SCSIVolumeAttachment) DeepCopy() *SCSIVolumeAttachment {
	if in == nil {
		return nil
	}
	out := new(SCSIVolumeAttachment)
	in.DeepCopyInto(out)
	return out
}
