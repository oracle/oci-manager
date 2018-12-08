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

// NodePool names
const (
	NodePoolKind           = "NodePool"
	NodePoolResourcePlural = "nodepools"
	NodePoolControllerName = "nodepools"
)

// NodePoolValidation describes the nodePool validation schema
var NodePoolValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"compartmentRef", "clusterRef", "subnetRefs", "kubernetesVersion"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"compartmentRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"clusterRef": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"subnetRefs": {
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

// NodePool describes a nodePool
type NodePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              NodePoolSpec   `json:"spec"`
	Status            NodePoolStatus `json:"status,omitempty"`
}

// NodePoolSpec describes a nodePool spec
type NodePoolSpec struct {
	CompartmentRef string `json:"compartmentRef"`
	ClusterRef     string `json:"clusterRef"`

	// The references of the subnets in which to place nodes for this node pool.
	SubnetRefs []string `mandatory:"true" json:"subnetRefs"`

	// The version of Kubernetes to install on the nodes in the node pool.
	KubernetesVersion *string `mandatory:"true" json:"kubernetesVersion"`

	// The name of the image running on the nodes in the node pool.
	NodeImageName *string `mandatory:"true" json:"nodeImageName"`

	// The name of the node shape of the nodes in the node pool.
	NodeShape *string `mandatory:"true" json:"nodeShape"`

	// A list of key/value pairs to add to nodes after they join the Kubernetes cluster.
	InitialNodeLabels []ocice.KeyValue `mandatory:"false" json:"initialNodeLabels"`

	// The SSH public key to add to each node in the node pool.
	SshPublicKey *string `mandatory:"false" json:"sshPublicKey"`

	// The number of nodes to create in each subnet.
	QuantityPerSubnet *int `mandatory:"false" json:"quantityPerSubnet"`

	common.Dependency
}

// NodePoolStatus describes a nodePool status
type NodePoolStatus struct {
	common.ResourceStatus
	Resource *NodePoolResource `json:"resource,omitempty"`

	WorkRequestId     *string                      `json:"workRequestId,omitempty"`
	WorkRequestStatus *ocice.WorkRequestStatusEnum `json:"workRequestStatus,omitempty"`
}

// NodePoolResource describes a nodePool resource from oci
type NodePoolResource struct {
	*ocice.NodePool
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodePoolList is a list of NodePool items
type NodePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []NodePool `json:"items"`
}

// IsResource returns true if there is an oci id, otherwise false
func (s *NodePool) IsResource() bool {
	if s.GetResourceID() == "" {
		return false
	}
	return true
}

// GetResourceID returns the oci id of the nodePool
func (s *NodePool) GetResourceID() string {
	var id string
	if s.Status.Resource != nil && s.Status.Resource.Name != nil {
		id = *s.Status.Resource.Name
	}
	return id
}

// GetResourcePlural returns the plural name of the nodePool type
func (s *NodePool) GetResourcePlural() string {
	return NodePoolResourcePlural
}

// GetGroupVersionResource returns the group version of the nodePool type
func (s *NodePool) GetGroupVersionResource() schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(NodePoolResourcePlural)
}

// SetResource sets the resource in nodePool status
func (s *NodePool) SetResource(r *ocice.NodePool) *NodePool {
	if r != nil {
		s.Status.Resource = &NodePoolResource{r}
	}
	return s
}

// GetResourceState returns the current state of the iresource
func (s *NodePool) GetResourceState() common.ResourceState {
	return s.Status.State
}

// AddDependent adds a nodePool dependent
func (s *NodePool) AddDependent(kind string, obj runtime.Object) error {
	return s.Status.AddDependent(kind, obj)
}

// RemoveDependent removes a nodePool dependent
func (s *NodePool) RemoveDependent(kind string, obj runtime.Object) error {
	return s.Status.RemoveDependent(kind, obj)
}

// IsDependentRegistered returns true if the nodePool dependent is registered
func (s *NodePool) IsDependentRegistered(kind string, obj runtime.Object) (bool, error) {
	return s.Status.IsDependentRegistered(kind, obj)
}

// DeepCopy the nodePool spec
func (in *NodePoolSpec) DeepCopy() *NodePoolSpec {
	if in == nil {
		return nil
	}
	out := in
	return out
}

// DeepCopy the backed oci resource
func (in *NodePoolResource) DeepCopy() (out *NodePoolResource) {
	if in == nil {
		return nil
	}
	out = in
	return
}
