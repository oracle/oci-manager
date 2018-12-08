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

package core

import (
	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/golang/glog"
	ociidv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	kubecommon "github.com/oracle/oci-manager/pkg/controller/oci/kubernetes/common"
)

const (
	compartmentLabel = "oci-compartment"
	controllerName   = "namespace"
	group            = ""
	resourcePlural   = "namespaces"
	version          = "v1"
)

func init() {
	kubecommon.RegisterKubernetesType(
		group,
		version,
		resourcePlural,
		controllerName,
		&corev1.Namespace{},
		NewNamespaceAdapter)
}

// NamespaceAdapter implements the adapter interface for Namespace resource
type NamespaceAdapter struct {
	clientset  versioned.Interface
	kubeclient kubernetes.Interface
	kind       string
}

// NewNamespaceAdapter creates a new adapter for Namespace resource
func NewNamespaceAdapter(
	clientset versioned.Interface,
	kubeclient kubernetes.Interface,
	adapterSpecificArgs map[string]interface{}) kubecommon.KubernetesTypeAdapter {

	na := NamespaceAdapter{
		clientset:  clientset,
		kubeclient: kubeclient,
	}
	return &na
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *NamespaceAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return corev1.SchemeGroupVersion.WithResource(resourcePlural)
}

// ObjectType returns the Namespace type for this adapter
func (a *NamespaceAdapter) ObjectType() runtime.Object {
	return &corev1.Namespace{}
}

// Kind returns the string type for this adapter
func (a *NamespaceAdapter) Kind() string {
	return controllerName
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *NamespaceAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*corev1.Namespace)
	return ok
}

// Copy returns a copy of a Namespace object
func (a *NamespaceAdapter) Copy(obj runtime.Object) runtime.Object {
	Namespace := obj.(*corev1.Namespace)
	return Namespace.DeepCopyObject()
}

// Equivalent checks if two Namespace objects are the same
func (a *NamespaceAdapter) IsCompliant(obj runtime.Object) bool {
	ns := obj.(*corev1.Namespace)

	if val, ok := ns.Labels[compartmentLabel]; ok {
		compartment, err := a.clientset.OciidentityV1alpha1().Compartments(ns.Name).Get(ns.Name, metav1.GetOptions{})

		if val == "true" {
			if err != nil {
				return false
			}
			return true

		} else if val == "false" {
			if compartment != nil {
				return false
			}
			return true

		} else {
			glog.Errorf("invalid %s value: %s", compartmentLabel, val)
		}

	}

	return true
}

// Id returns the name of the compartment if label exists to create it ... when empty the controller will call Create
func (a *NamespaceAdapter) Id(obj runtime.Object) string {
	ns := obj.(*corev1.Namespace)

	if val, ok := ns.Labels[compartmentLabel]; ok {
		if val == "true" {
			existingCompartment, _ := a.clientset.OciidentityV1alpha1().Compartments(ns.Name).Get(ns.Name, metav1.GetOptions{})
			if existingCompartment == nil {
				return ""
			} else {
				return existingCompartment.Name
			}
		}
	}
	return ns.Name
}

// ObjectMeta returns the object meta struct from the Namespace object
func (a *NamespaceAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*corev1.Namespace).ObjectMeta
}

// CreateObject creates the Namespace object
func (a *NamespaceAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {

	// return nil to prevent namespace update
	return nil, nil
}

// UpdateObject updates the Namespace object
func (a *NamespaceAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {

	// cannot update namespace
	return nil, nil
}

// DeleteObject deletes the Namespace object
func (a *NamespaceAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var object = obj.(*corev1.Namespace)
	return a.clientset.OciidentityV1alpha1().Compartments(object.ObjectMeta.Namespace).Delete(object.Name, options)
}

// Create creates the Namespace resource in oci
func (a *NamespaceAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	var (
		ns  = obj.(*corev1.Namespace)
		err error
	)

	existingCompartment, err := a.clientset.OciidentityV1alpha1().Compartments(ns.Name).Get(ns.Name, metav1.GetOptions{})
	if err != nil {
		glog.Infof("err getting Compartment: %v", err)
	}
	if apierrors.IsNotFound(err) {
		glog.Infof("create compartment for namespace: %s", ns.Name)
		comp := &ociidv1alpha1.Compartment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ns.Name,
				Namespace: ns.Name,
			},
			Spec: ociidv1alpha1.CompartmentSpec{
				Description: ns.Name + " created from oci-manager namespace controller",
			},
		}

		existingCompartment, err = a.clientset.OciidentityV1alpha1().Compartments(ns.Name).Create(comp)
		if err != nil {
			glog.Infof("err creating compartment: %v", err)
		}
	}
	if existingCompartment == nil {
		glog.Errorf("err creating compartment: %v", err)
		return nil, nil
	}

	return ns, nil
}

// Update updates the Namespace resource in oci
func (a *NamespaceAdapter) Update(obj runtime.Object) (runtime.Object, error) {

	var ns = obj.(*corev1.Namespace)

	glog.Infof("update ns: %s", ns.Name)

	if val, ok := ns.Labels[compartmentLabel]; ok {
		_, err := a.clientset.OciidentityV1alpha1().Compartments(ns.Name).Get(ns.Name, metav1.GetOptions{})
		if val == "true" && err != nil {
			return a.Create(ns)
		} else if val == "false" && err == nil {
			err := a.clientset.OciidentityV1alpha1().Compartments(ns.Name).Delete(ns.Name, &metav1.DeleteOptions{})
			if err != nil {
				glog.Errorf("error deleting compartment: %s err: %v", ns.Name, err)
			}
		}
	}

	return ns, nil
}

// Delete deletes the Namespace resource in oci
func (a *NamespaceAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	var ns = obj.(*corev1.Namespace)
	return ns, nil
}

// Get retrieves the Namespace resource from oci
func (a *NamespaceAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	var ns = obj.(*corev1.Namespace)
	return ns, nil
}
