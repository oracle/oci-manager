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
	"github.com/oracle/oci-manager/pkg/apis/ocilb.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LoadBalancer returns the load balancer object for the receiving oci resource
func LoadBalancer(clientset versioned.Interface, ns, name string) (lb *v1alpha1.LoadBalancer, err error) {
	lb, err = clientset.OcilbV1alpha1().LoadBalancers(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return lb, err
	}

	return lb, nil
}

// LoadBalancerId returns the oci id of the load balancer for the receiving oci resource
func LoadBalancerId(clientset versioned.Interface, ns, name string) (id string, err error) {
	lb, err := clientset.OcilbV1alpha1().LoadBalancers(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return id, err
	}
	if lb.Status.Resource == nil || *lb.Status.Resource.Id == "" {
		return id, errors.New("LoadBalancer resource is not created")
	}
	return *lb.Status.Resource.Id, nil
}

// Backend returns the backend object for the receiving oci resource
func Backend(clientset versioned.Interface, ns, name string) (be *v1alpha1.Backend, err error) {
	be, err = clientset.OcilbV1alpha1().Backends(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return be, err
	}

	if be.Status.Resource == nil {
		return be, errors.New("Backend resource is not created")
	}
	return be, nil
}

// BackendSet returns the backend set object for the receiving oci resource
func BackendSet(clientset versioned.Interface, ns, name string) (bes *v1alpha1.BackendSet, err error) {
	bes, err = clientset.OcilbV1alpha1().BackendSets(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return bes, err
	}

	return bes, nil
}

// Listener returns the listener object for the receiving oci resource
func Listener(clientset versioned.Interface, ns, name string) (l *v1alpha1.Listener, err error) {
	l, err = clientset.OcilbV1alpha1().Listeners(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return l, err
	}

	if l.Status.Resource == nil {
		return l, errors.New("Listener resource is not created")
	}
	return l, nil
}
