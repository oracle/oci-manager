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
	"errors"
	"github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Compartment returns the compartment object for the receving oci resource
func Compartment(clientset versioned.Interface, ns, name string) (compartment *v1alpha1.Compartment, err error) {
	compartment, err = clientset.OciidentityV1alpha1().Compartments(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return compartment, err
	}

	if compartment.Status.Resource == nil || *compartment.Status.Resource.Id == "" {
		return compartment, errors.New("Compartment resource is not created")
	}
	return compartment, nil
}

// CompartmentId returns the oci id of the compartment for the receving oci resource
func CompartmentId(clientset versioned.Interface, ns, name string) (id string, err error) {
	compartment, err := clientset.OciidentityV1alpha1().Compartments(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return id, err
	}
	if compartment.Status.Resource == nil || *compartment.Status.Resource.Id == "" {
		return id, errors.New("Compartment resource is not created")
	}
	return *compartment.Status.Resource.Id, nil
}
