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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	common "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

const (
	NetworkKind           = "Network"
	NetworkResourcePlural = "networks"
	NetworkControllerName = "networks"
)

// NetworkValidation describes the network validation schema
var NetworkValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"cirdBlock": {
						Type:    common.ValidationTypeString,
						Pattern: common.CidrValidationRegex,
					},
				},
			},
		},
	},
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//Network describes network
type Network struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              NetworkSpec   `json:"spec"`
	Status            NetworkStatus `json:"status,omitempty"`
}

type NetworkSpec struct {
	CidrBlock string `json:"cidrBlock"`
}

type NetworkStatus struct {
	OperatorStatus

	// subnet allocation map
	// key: compute,lb or subnet name, value: floor to tens of subnet octet. ie:
	// 10, 20 ... 250 ...allows for 25 compute on a network,
	// with 9 az (subnet x.y.n1-n9.0/24) per compute ... max 57150 instances per network
	SubnetAllocationMap map[string]int
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//NetworkList is a list of Network resources
type NetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Network `json:"items"`
}

type Subnet struct {
	Private bool `json:"private"`
}
