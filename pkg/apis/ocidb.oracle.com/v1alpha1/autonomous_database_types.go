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
	ocidb "github.com/oracle/oci-go-sdk/database"
	common "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// AutonomousDatabase names
const (
	AutonomousDatabaseKind           = "AutonomousDatabase"
	AutonomousDatabaseResourcePlural = "autonomousdatabases"
	AutonomousDatabaseControllerName = "autonomousdatabases"
)

// AutonomousDatabaseValidation describes the AutonomousDatabase validation schema
var AutonomousDatabaseValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"compartmentRef", "cpuCoreCount", "dataStorageSizeInTBs"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"compartmentRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"cpuCoreCount": {
						Type: common.ValidationTypeInteger,
					},
					"dataStorageSizeInTBs": {
						Type: common.ValidationTypeInteger,
					},
				},
			},
		},
	},
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AutonomousDatabase describes a AutonomousDatabase
type AutonomousDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              AutonomousDatabaseSpec   `json:"spec"`
	Status            AutonomousDatabaseStatus `json:"status,omitempty"`
}

// AutonomousDatabaseSpec describes a AutonomousDatabase spec
type AutonomousDatabaseSpec struct {
	CompartmentRef string `json:"compartmentRef"`

	// The number of CPU cores to be made available to the database.
	CpuCoreCount *int `mandatory:"true" json:"cpuCoreCount"`

	// The quantity of data in the database, in terabytes.
	DataStorageSizeInTBs *int `mandatory:"true" json:"dataStorageSizeInTBs"`

	// The user-friendly name for the Autonomous Database. The name does not have to be unique.
	DisplayName string `mandatory:"false" json:"displayName"`

	// The Oracle license model that applies to the Oracle Autonomous Database. The default is BRING_YOUR_OWN_LICENSE.
	LicenseModel ocidb.AutonomousDatabaseLicenseModelEnum `mandatory:"false" json:"licenseModel,omitempty"`

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

// AutonomousDatabaseStatus describes a AutonomousDatabase status
type AutonomousDatabaseStatus struct {
	common.ResourceStatus

	Resource *AutonomousDatabaseResource `json:"resource,omitempty"`
}

// AutonomousDatabaseResource describes a AutonomousDatabase resource from oci
type AutonomousDatabaseResource struct {
	*ocidb.AutonomousDatabase
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AutonomousDatabaseList is a list of AutonomousDatabase items
type AutonomousDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []AutonomousDatabase `json:"items"`
}

// IsResource returns true if there is an oci id and it's in a running state, otherwise false
func (s *AutonomousDatabase) IsResource() bool {
	if s.GetResourceID() != "" && s.GetResourceLifecycleState() == string(ocidb.AutonomousDatabaseLifecycleStateAvailable) {
		return true
	}
	return false
}

// GetResourceLifecycleState returns the current state of the instance
func (s *AutonomousDatabase) GetResourceLifecycleState() string {
	var state string
	if s.Status.Resource != nil {
		state = string(s.Status.Resource.LifecycleState)
	}
	return state
}

// GetResourceID returns the oci id of the AutonomousDatabase
func (s *AutonomousDatabase) GetResourceID() string {
	if s.Status.Resource != nil && s.Status.Resource.Id != nil {
		return *s.Status.Resource.Id
	}
	return ""
}

// GetResourcePlural returns the plural name of the AutonomousDatabase type
func (s *AutonomousDatabase) GetResourcePlural() string {
	return AutonomousDatabaseResourcePlural
}

// GetGroupVersionResource returns the group version of the AutonomousDatabase type
func (s *AutonomousDatabase) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(AutonomousDatabaseResourcePlural)
}

// SetResource sets the resource in AutonomousDatabase status
func (s *AutonomousDatabase) SetResource(r *ocidb.AutonomousDatabase) *AutonomousDatabase {
	if r != nil {
		s.Status.Resource = &AutonomousDatabaseResource{r}
	}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *AutonomousDatabase) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds a AutonomousDatabase dependent
func (s *AutonomousDatabase) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes a AutonomousDatabase dependent
func (s *AutonomousDatabase) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the AutonomousDatabase dependent is registered
func (s *AutonomousDatabase) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the AutonomousDatabase spec
func (in *AutonomousDatabaseSpec) DeepCopy() *AutonomousDatabaseSpec {
	if in == nil {
		return nil
	}
	out := in
	return out
}

// DeepCopy the backed oci resource
func (in *AutonomousDatabaseResource) DeepCopy() (out *AutonomousDatabaseResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
