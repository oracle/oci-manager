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
	ComputeKind           = "Compute"
	ComputeResourcePlural = "computes"
	ComputeControllerName = "computes"
)

// ComputeValidation describes the compute validation schema
var ComputeValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"network"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"network": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"replicas": {
						Type: common.ValidationTypeInteger,
					},
					"minAvailabilityZones": {
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

//Compute describes Compute
type Compute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ComputeSpec   `json:"spec"`
	Status            ComputeStatus `json:"status,omitempty"`
}

type ComputeSpec struct {
	// name of the network to place the compute
	Network string `json:"network"`

	// number of instances
	Replicas int `json:"replicas,omitempty"`

	// minimum availability zone count
	MinAvailabiliyZones int `json:"minAvailabilityZones,omitempty"`

	SecuritySelector map[string]string `json:"securitySelector,omitempty"`

	// attributes that each instance will use to construct its model
	Template Template `json:"template,omitempty"`

	Env       []apiv1.EnvVar             `json:"env,omitempty"`
	Resources apiv1.ResourceRequirements `json:"resources,omitempty"`
}

// attributes that each instance will use to constuct its model
type Template struct {
	// os type and version eg centos 7
	OsType    string `json:"osType,omitempty"`
	OsVersion string `json:"osVersion,omitempty"`

	// reuse kubernetes ResourceRequirements model - https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
	Resources apiv1.ResourceRequirements `json:"resources,omitempty"`

	// ssh public keys to allow remote access
	SshKeys []string `json:"sshKeys,omitempty"`

	// user-data for cloud-init
	UserData UserData `json:"userData,omitempty"`

	Volumes []Volume `json:"volume,omitempty"`
}

type UserData struct {
	Shellscript    string `json:"shellscript,omitempty"`
	CloudConfig    string `json:"cloud-config,omitempty"`
	IncludeURL     string `json:"include-url,omitempty"`
	IncludeURLOnce string `json:"include-url-once,omitempty"`
	PartHandler    string `json:"part-handler,omitempty"`
}

type Volume struct {
	Size string `json:"size"`
	Type string `json:"type"`

	// fstab related
	MountPoint     string `json:""`
	FilesystemType string `json:"filesystemType"`
	Options        map[string]string
}

type ComputeStatus struct {
	OperatorStatus

	// one-time randomized array of availability zones/domains for even distribution
	AvailabilityZones []string `json:"availabilityZones"`

	Replicas            int `json:"replicas"`
	ReadyReplicas       int `json:"readyReplicas"`
	AvailableReplicas   int `json:"availableReplicas"`
	UnavailableReplicas int `json:"unavailableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//ComputeList is a list of Compute resources
type ComputeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Compute `json:"items"`
}
