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

// Listener names
const (
	ListenerKind           = "Listener"
	ListenerResourcePlural = "listeners"
	ListenerControllerName = "listeners"
)

// ListenerValidation describes the listener validation schema
var ListenerValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"loadBalancerRef", "defaultBackendSetName", "port", "protocol"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"certificateRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.NoOrAnyStringValidationRegex,
					},
					"loadBalancerRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"defaultBackendSetName": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"port": {
						Type:    common.ValidationTypeInteger,
						Minimum: &minTCPPort,
						Maximum: &maxTCPPort,
					},
					"protocol": {
						Type:    common.ValidationTypeString,
						Pattern: common.LoadBalancerProtocolRegex,
					},
				},
			},
		},
	},
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Listener describes a listener
type Listener struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ListenerSpec   `json:"spec"`
	Status            ListenerStatus `json:"status,omitempty"`
}

// ListenerSpec describes a listener spec
type ListenerSpec struct {
	CertificateRef  string `json:"certificateRef"`
	LoadBalancerRef string `json:"loadBalancerRef"`

	DefaultBackendSetName string `header:"-" url:"-" json:"defaultBackendSetName"`
	Port                  int    `header:"-" url:"-" json:"port"`
	Protocol              string `header:"-" url:"-" json:"protocol"`
	IdleTimeout           int64  `header:"-" url:"-" json:"idleTimeout"`
	PathRouteSetName      string `header:"-" url:"-" json:"pathRouteSetName"`

	common.Dependency
}

// ListenerStatus describes a listener status
type ListenerStatus struct {
	common.ResourceStatus
	LoadBalancerId *string

	WorkRequestId     *string                              `json:"workRequestId,omitempty"`
	WorkRequestStatus *ocilb.WorkRequestLifecycleStateEnum `json:"workRequestStatus,omitempty"`

	Resource *ListenerResource `json:"resource,omitempty"`
}

// ListenerResource describes a listener resource from oci
type ListenerResource struct {
	*ocilb.Listener
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ListenerList is a list of Listener items
type ListenerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Listener `json:"items"`
}

// IsResource returns true if there is an oci id, otherwise false
func (s *Listener) IsResource() bool {
	if s.GetResourceID() == "" {
		return false
	}
	return true
}

// GetResourceID returns the oci id of the listener
func (s *Listener) GetResourceID() string {
	var id string
	if s.Status.Resource != nil {
		id = *s.Status.Resource.Name
	}
	return id
}

// GetResourcePlural returns the plural name of the listener type
func (s *Listener) GetResourcePlural() string {
	return ListenerResourcePlural
}

// GetGroupVersionResource returns the group version of the listener type
func (s *Listener) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(ListenerResourcePlural)
}

// SetResource sets the resource in listener status
func (s *Listener) SetResource(r *ocilb.Listener) *Listener {
	if r != nil {
		s.Status.Resource = &ListenerResource{r}
	}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *Listener) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds a listener dependent
func (s *Listener) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes a listener dependent
func (s *Listener) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the listener dependent is registered
func (s *Listener) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the listener spec
func (in *ListenerSpec) DeepCopy() *ListenerSpec {
	if in == nil {
		return nil
	}
	out := in
	return out
}

// DeepCopy the listener oci resource
func (in *ListenerResource) DeepCopy() (out *ListenerResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
