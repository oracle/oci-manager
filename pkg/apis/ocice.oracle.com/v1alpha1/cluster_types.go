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
	ocice "github.com/oracle/oci-go-sdk/containerengine"
	common "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Cluster names
const (
	ClusterKind           = "Cluster"
	ClusterResourcePlural = "clusters"
	ClusterControllerName = "clusters"
)

// ClusterValidation describes the cluster validation schema
var ClusterValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"compartmentRef", "vcnRef", "serviceLbSubnetRefs", "kubernetesVersion"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"compartmentRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"vcnRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"serviceLbSubnetRefs": {
						Type: common.ValidationTypeArray,
					},
					"kubernetesVersion": {
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

// Cluster describes a cluster
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ClusterSpec   `json:"spec"`
	Status            ClusterStatus `json:"status,omitempty"`
}

// ClusterSpec describes a cluster spec
type ClusterSpec struct {
	CompartmentRef      string   `json:"compartmentRef"`
	ServiceLbSubnetRefs []string `json:"serviceLbSubnetRefs"`
	VcnRef              string   `json:"vcnRef"`

	KubernetesVersion *string                     `mandatory:"true" json:"kubernetesVersion"`
	Options           *ocice.ClusterCreateOptions `mandatory:"false" json:"options"`

	common.Dependency
}

// ClusterStatus describes a cluster status
type ClusterStatus struct {
	common.ResourceStatus

	KubeConfig *string          `json:"kubeconfig,omitempty"`
	Resource   *ClusterResource `json:"resource,omitempty"`

	WorkRequestId     *string                      `json:"workRequestId,omitempty"`
	WorkRequestStatus *ocice.WorkRequestStatusEnum `json:"workRequestStatus,omitempty"`
}

// ClusterResource describes a cluster resource from oci
type ClusterResource struct {
	*ocice.Cluster
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterList is a list of Cluster items
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Cluster `json:"items"`
}

// IsResource returns true if there is an oci id, otherwise false
func (s *Cluster) IsResource() bool {
	if s.GetResourceID() == "" {
		return false
	}
	return true
}

// GetResourceID returns the oci id of the cluster
func (s *Cluster) GetResourceID() string {
	var id string
	if s.Status.Resource != nil && s.Status.Resource.Id != nil {
		id = *s.Status.Resource.Id
	}
	return id
}

// GetResourcePlural returns the plural name of the cluster type
func (s *Cluster) GetResourcePlural() string {
	return ClusterResourcePlural
}

// GetGroupVersionResource returns the group version of the cluster type
func (s *Cluster) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(ClusterResourcePlural)
}

// SetResource sets the resource in cluster status
func (s *Cluster) SetResource(r *ocice.Cluster) *Cluster {
	if r != nil {
		s.Status.Resource = &ClusterResource{r}
	}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *Cluster) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds a cluster dependent
func (s *Cluster) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes a cluster dependent
func (s *Cluster) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the cluster dependent is registered
func (s *Cluster) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the cluster spec
func (in *ClusterSpec) DeepCopy() *ClusterSpec {
	if in == nil {
		return nil
	}
	out := in
	return out
}

// DeepCopy the backed oci resource
func (in *ClusterResource) DeepCopy() (out *ClusterResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
