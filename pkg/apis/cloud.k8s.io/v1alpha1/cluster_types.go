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
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	common "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

const (
	ClusterKind           = "Cluster"
	ClusterResourcePlural = "clusters"
	ClusterControllerName = "clusters"
)

// ClusterValidation describes the Cluster validation schema
var ClusterValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
		},
	},
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//Cluster describes Cluster
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ClusterSpec   `json:"spec"`
	Status            ClusterStatus `json:"status,omitempty"`
}

type ClusterSpec struct {
	// to managed service vs kubeadm
	IsManaged bool `json:"isManaged"`

	Version string `json:"version,omitempty"`

	Master ComputeTemplate `json:"master"`
	Worker ComputeTemplate `json:"worker"`

	// for kubeadm/non-managed service to remotely create user/kubeconfig
	CA    CertificateAuthority `json:"ca,omitempty"`
	Token string               `json:"token"`

	Env       []apiv1.EnvVar             `json:"env,omitempty"`
	Resources apiv1.ResourceRequirements `json:"resources,omitempty"`
}

type CertificateAuthority struct {
	Certificate string `json:"certificate"`
	Key         string `json:"key"`
}

type ComputeTemplate struct {
	// number of instances
	Replicas int `json:"replicas,omitempty"`

	// template passed down to compute
	Template Template `json:"template,omitempty"`
}

type ClusterStatus struct {
	AvailabilityZones []string `json:"availabilityZones"`

	OperatorStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//ClusterList is a list of Cluster resources
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Cluster `json:"items"`
}
