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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResourceTypeAdapter defines operations for interacting with a
// federated type.  Code written to this interface can then target any
// type for which an implementation of this interface exists.
type KubernetesTypeAdapter interface {
	ObjectType() runtime.Object
	GroupVersionWithResource() schema.GroupVersionResource

	IsExpectedType(obj interface{}) bool
	IsCompliant(obj runtime.Object) bool
	Id(obj runtime.Object) string
	//Key(obj runtime.Object) objectclient.Key
	ObjectMeta(obj runtime.Object) *metav1.ObjectMeta
	Kind() string
	Copy(obj runtime.Object) runtime.Object

	// Operations to update resources
	Create(obj runtime.Object) (runtime.Object, error)
	Get(obj runtime.Object) (runtime.Object, error)
	Update(obj runtime.Object) (runtime.Object, error)
	Delete(obj runtime.Object) (runtime.Object, error)

	// Operations to handle resource updates
	CreateObject(obj runtime.Object) (runtime.Object, error)
	UpdateObject(obj runtime.Object) (runtime.Object, error)
	DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error
}

// AdapterFactory defines the function signature for factory methods
// that create instances of ResourceTypeAdapter.  Such methods should
// be registered with RegisterAdapterFactory to ensure the type
// adapter is discoverable.
type AdapterFactory func(
	clientset versioned.Interface,
	kubeclient kubernetes.Interface,
	adapterSpecificArgs map[string]interface{}) KubernetesTypeAdapter
