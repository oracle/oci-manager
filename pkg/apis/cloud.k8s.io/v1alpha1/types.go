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
)

type OperatorStatus struct {
	State      OperatorState       `json:"state,omitempty"`
	Message    string              `json:"message,omitempty"`
	Conditions []OperatorCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

type OperatorState string

const (
	OperatorStatePending   OperatorState = "Pending"
	OperatorStateCreated   OperatorState = "Created"
	OperatorStateProcessed OperatorState = "Processed"
	OperatorStateError     OperatorState = "Error"
)

type OperatorConditionType string

type OperatorCondition struct {
	Type               OperatorConditionType `json:"type"`
	Reason             string                `json:"reason,omitempty"`
	LastTransitiontime metav1.Time           `json:"lastTransitionTime,omitempty"`
}
