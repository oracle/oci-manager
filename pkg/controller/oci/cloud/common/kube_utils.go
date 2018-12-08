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

package common

import (
	cloudv1alpha1 "github.com/oracle/oci-manager/pkg/apis/cloud.k8s.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//Create OwnerReference
func CreateControllerRef(controller metav1.Object, groupVersionWithKind schema.GroupVersionKind) *metav1.OwnerReference {
	boolPtr := func(b bool) *bool { return &b }
	controllerRef := &metav1.OwnerReference{
		APIVersion:         groupVersionWithKind.String(),
		Kind:               groupVersionWithKind.Kind,
		Name:               controller.GetName(),
		UID:                controller.GetUID(),
		BlockOwnerDeletion: boolPtr(true),
		Controller:         boolPtr(true),
	}
	return controllerRef
}

//Retrieves OwnerReference from the object
func GetControllerOf(controllee metav1.Object) *metav1.OwnerReference {
	ownerRefs := controllee.GetOwnerReferences()
	for i := range ownerRefs {
		owner := &ownerRefs[i]
		if owner.Controller != nil && *owner.Controller == true {
			return owner
		}
	}
	return nil
}

//Get condition (status) from object by the type
func GetCondition(status *cloudv1alpha1.OperatorStatus, conditionType cloudv1alpha1.OperatorConditionType) *cloudv1alpha1.OperatorCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == conditionType {
			return &c
		}
	}
	return nil
}

//Set the condition (status) on the object by the type
func SetCondition(status *cloudv1alpha1.OperatorStatus, conditionType cloudv1alpha1.OperatorConditionType, reason string) {
	currentCondition := GetCondition(status, conditionType)
	if currentCondition != nil && currentCondition.Reason == reason {
		return
	}
	newConditions := FilteredConditions(status.Conditions, conditionType)
	condition := cloudv1alpha1.OperatorCondition{
		Type:               conditionType,
		Reason:             reason,
		LastTransitiontime: metav1.Now(),
	}
	status.Conditions = append(newConditions, condition)
}

// FilteredConditions returns a new slice of operator conditions without conditions with the provided type.
func FilteredConditions(conditions []cloudv1alpha1.OperatorCondition, conditionType cloudv1alpha1.OperatorConditionType) []cloudv1alpha1.OperatorCondition {
	var newConditions []cloudv1alpha1.OperatorCondition
	for _, c := range conditions {
		if c.Type == conditionType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}

// RemoveCondition removes the condition with the provided type from the operator status.
func RemoveCondition(status *cloudv1alpha1.OperatorStatus, conditionType cloudv1alpha1.OperatorConditionType) {
	status.Conditions = FilteredConditions(status.Conditions, conditionType)
}
