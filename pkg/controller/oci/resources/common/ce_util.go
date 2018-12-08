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
	"github.com/oracle/oci-manager/pkg/apis/ocice.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Cluster returns the cluster object for the receiving oci resource
func Cluster(clientset versioned.Interface, ns, name string) (cluster *v1alpha1.Cluster, err error) {

	cluster, err = clientset.OciceV1alpha1().Clusters(ns).Get(name, metav1.GetOptions{})

	if err != nil {
		return cluster, err
	}
	return cluster, nil
}

// NodePool returns the nodepool object for the receiving oci resource
func NodePool(clientset versioned.Interface, ns, name string) (nodepool *v1alpha1.NodePool, err error) {

	nodepool, err = clientset.OciceV1alpha1().NodePools(ns).Get(name, metav1.GetOptions{})

	if err != nil {
		return nodepool, err
	}
	return nodepool, nil
}
