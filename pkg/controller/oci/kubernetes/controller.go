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

package kubernetes

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
	clientset "github.com/oracle/oci-manager/pkg/client/clientset/versioned"

	"k8s.io/client-go/informers"

	kubecommon "github.com/oracle/oci-manager/pkg/controller/oci/kubernetes/common"
)

// OciGroupName constant is used for finalizers string
const (
	OciGroupName = "oci.oracle.com"
)

// Controller of resource create/update/delete events
type Controller struct {
	queue    workqueue.RateLimitingInterface
	queueMap map[string]workqueue.RateLimitingInterface
	informer cache.SharedInformer
	adapter  kubecommon.KubernetesTypeAdapter
	factory  informers.SharedInformerFactory
}

// Start a new controller for a type adapter
func Start(
	clientset clientset.Interface,
	kubeclient kubernetes.Interface,
	watchType runtime.Object,
	adapterFactory kubecommon.AdapterFactory,
	adapterSpecificArgs map[string]interface{},
	informerFactory informers.SharedInformerFactory,
	stopChan <-chan struct{},
	queueMap map[string]workqueue.RateLimitingInterface,
) *Controller {
	adapter := adapterFactory(clientset, kubeclient, adapterSpecificArgs)
	controller := New(adapter, watchType, informerFactory, queueMap)
	controller.Run(stopChan)
	return controller
}

// New initializes a controller object
func New(adapter kubecommon.KubernetesTypeAdapter,
	objectType runtime.Object,
	informerFactory informers.SharedInformerFactory,
	queueMap map[string]workqueue.RateLimitingInterface) *Controller {

	kind := reflect.TypeOf(objectType).String()

	c := &Controller{
		adapter:  adapter,
		factory:  informerFactory,
		queue:    queueMap[kind],
		queueMap: queueMap,
	}

	// using explicit group version kind via adapter due to these are empty:
	// 	glog.Infof("Group: %s Version: %s Resource: %s",  objectType.GetObjectKind().GroupVersionKind().Group,
	//		objectType.GetObjectKind().GroupVersionKind().Version,  objectType.GetObjectKind().GroupVersionKind().Kind)

	informer, err := informerFactory.ForResource(adapter.GroupVersionWithResource())
	if err != nil {
		glog.Infof("could not create")
	}
	c.informer = informer.Informer()

	c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(key)
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(cur)
			if err == nil {
				c.queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(key)
			}
		},
	})

	return c
}

// Run turns on the controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	// TODO add multiple workers for performance
	go wait.Until(c.worker, time.Second, stopCh)

	go func() {
		<-stopCh
		c.queue.ShutDown()
	}()

}

// HasSynced is required for the cache.Controller interface.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *Controller) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
}

// Worker start a loop for a single worker routine
func (c *Controller) worker() {
	for c.dequeue() {
		// loop
	}
}

// Dequeue an event from the queue to process
func (c *Controller) dequeue() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	object, err, retry := c.reconcile(key.(string))

	if err == nil {
		if retry && c.queue.NumRequeues(key) <= 30 {
			glog.V(2).Infof("Resubmit for reconcile key: %v", key)
			c.queue.AddAfter(key, 5*time.Second)
			return true
		}
		// No error, no retries reset the ratelimit counters and update the object
		err := c.report(key.(string), object)
		if err != nil {
			c.queue.AddRateLimited(key)
		} else {
			c.queue.Forget(key)
		}
	} else if c.queue.NumRequeues(key) <= 5 {
		glog.Errorf("Error reconciling key %v - %v", key, err)
		c.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		c.report(key.(string), object)
		c.queue.Forget(key)
		utilruntime.HandleError(err)
	}

	return true
}

// Report object changes to the client to update/save
func (c *Controller) report(key string, object runtime.Object) error {
	if object != nil {
		kind := c.adapter.Kind()
		glog.V(1).Infof("Updating object %s  %s \n", kind, key)
		glog.V(4).Infof("Updating object %s  %s --- %#v\n", kind, key, object)
		if _, e := c.adapter.UpdateObject(object); e != nil {
			glog.Errorf("ERROR updating object kind %s and key %s: %#v\n", kind, key, e)
			return e
		}
	}
	return nil
}

// Reconcile the resource with the incoming object specification.
// This is the main func doing the work to manage the lifecycle of a resource
func (c *Controller) reconcile(key string) (reconciled runtime.Object, err error, retry bool) {

	kind := c.adapter.Kind()
	glog.V(2).Infof("Starting to reconcile %v %v\n", kind, key)
	startTime := time.Now()
	defer glog.V(2).Infof("Finished reconciling %v %v (duration: %v)\n", kind, key, time.Now().Sub(startTime))

	obj, exists, err := c.informer.GetStore().GetByKey(key)

	if err != nil {
		return nil, fmt.Errorf("ERROR fetching object with key %s from store: %v", key, err), false
	}

	if !exists {
		return nil, nil, false
	}

	// Copy the object first instead of updating the one in the cache
	source := obj.(runtime.Object)
	object := source.DeepCopyObject()
	objectmeta := c.adapter.ObjectMeta(object)

	if objectmeta.DeletionTimestamp != nil {
		return object, nil, false
	}

	// Create
	// The Id indicates if the resource is already created or it's pending.
	// After setting the finalizer we return to start the reconcile processing
	// from the beginning since the object was updated. It should be skipped on
	// the next loop and proceed with the Get below to validate the create
	if c.adapter.Id(object) == "" {
		glog.V(1).Infof("Creating resource %s  %s \n", kind, key)
		glog.V(5).Infof("Creating resource %s  %s --- %#v\n", kind, key, object)
		object, err = c.adapter.Create(object)
		if err != nil {
			glog.Errorf("ERROR creating resource kind %s and key %s: %#v\n", kind, key, err)
			return object, err, false
		}
		return object, nil, false
	}

	// Get
	// If the current state of the object is not pending a delete or create
	// we proceed with a Get call to fetch the remote resource so we can compare
	// to the current object
	object, err = c.adapter.Get(object)
	if err != nil {
		glog.Errorf("ERROR getting resource kind %s and key %s: %#v\n", kind, key, err)
		return object, err, false
	}

	// Update
	// If the object we got back from the Get call above differs from the incoming
	// object from queue we proceed with the resource update. The delta can be due
	// to a change in the spec of the object that needs to be reconciled with the
	// current resource or it could be because of some external change of the resource
	// that needs to be corrected since we are the source of truth for the resource

	if object != nil && !c.adapter.IsCompliant(object) {
		glog.V(1).Infof("Updating resource %s  %s %#v\n", kind, key, object)
		//this updates underlying oci resource not the crd
		// crd will get updated in report func
		object, err = c.adapter.Update(object)
		if err != nil {
			glog.Errorf("ERROR updating resource kind %s and key %s: %#v\n", kind, key, err)
			return object, err, false
		}
		return object, nil, false
	}

	// Finally if there is no change it's a no-op
	glog.V(4).Infof("Match resource %s  %s\n", kind, key)
	return nil, nil, false

}
