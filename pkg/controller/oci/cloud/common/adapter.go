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
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// ResourceTypeAdapter defines operations for interacting with a
// cloud workload type.  Code written to this interface can then target any
// type for which an implementation of this interface exists.
type CloudTypeAdapter interface {
	Kind() string
	Resource() string
	Equivalent(obj1, obj2 runtime.Object) bool
	GroupVersionWithResource() schema.GroupVersionResource
	Subscriptions() []schema.GroupVersionResource
	CallbackForResource(schema.GroupVersionResource) cache.ResourceEventHandlerFuncs
	SetLister(lister cache.GenericLister)
	SetQueue(queue workqueue.RateLimitingInterface)
	ObjectMeta(obj runtime.Object) *metav1.ObjectMeta
	Delete(obj runtime.Object) (runtime.Object, error)
	Update(obj runtime.Object) (runtime.Object, error)
	Reconcile(obj runtime.Object) (runtime.Object, error)
}

// AdapterFactory defines the function signature for factory methods
// that create instances of CloudTypeAdapter.  Such methods should
// be registered with RegisterAdapterFactory to ensure the type
// adapter is discoverable.
type AdapterFactory func(clientset versioned.Interface, kubeclient kubernetes.Interface) CloudTypeAdapter
