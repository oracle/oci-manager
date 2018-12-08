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

package cloud

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"time"

	cloudv1alpha1 "github.com/oracle/oci-manager/pkg/apis/cloud.k8s.io/v1alpha1"
	vclientset "github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	informers "github.com/oracle/oci-manager/pkg/client/informers/externalversions"
	cloudcommon "github.com/oracle/oci-manager/pkg/controller/oci/cloud/common"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/golang/glog"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	MAX_REQUEUES                   = 5
	CloudProivderAnnotation string = "cloud-provider"
	OCIProviderValue        string = "oci"
)

type Controller struct {
	queue    workqueue.RateLimitingInterface
	queueMap map[string]workqueue.RateLimitingInterface
	informer cache.SharedInformer
	adapter  cloudcommon.CloudTypeAdapter
	//cloudIfactory  informers.SharedInformerFactory
	//resourceIfactory  informers.SharedInformerFactory
}

func New(adapterFactory cloudcommon.AdapterFactory,
	clientSet vclientset.Interface,
	kubeclient kubernetes.Interface,
	cloudIFactory, resourceIFactory informers.SharedInformerFactory,
	queueMap map[string]workqueue.RateLimitingInterface) *Controller {

	adapter := adapterFactory(clientSet, kubeclient)

	c := &Controller{
		adapter:  adapter,
		queue:    queueMap[adapter.Kind()],
		queueMap: queueMap,
	}

	cloudInfomer, err := cloudIFactory.ForResource(adapter.GroupVersionWithResource())

	if err != nil {
		glog.Fatalf("Error building informer for resource: %s - %v", adapter.Resource(), err)
	}

	c.informer = cloudInfomer.Informer()
	c.adapter.SetLister(cloudInfomer.Lister())
	c.adapter.SetQueue(c.queue)

	c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queueUsingAnnotation(obj, key)
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(cur)
			if err == nil {
				c.queueUsingAnnotation(cur, key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queueUsingAnnotation(obj, key)
			}
		},
	})

	for _, subResource := range adapter.Subscriptions() {
		informer, err := resourceIFactory.ForResource(subResource)
		if err != nil {
			glog.Fatalf("Error building subscriber informer for resource: %s - %v", subResource, err)
		}
		informer.Informer().AddEventHandler(adapter.CallbackForResource(subResource))
	}

	return c

}

// cast to metav1 Object to get annotations and queue event if needed
func (c *Controller) queueUsingAnnotation(obj interface{}, key string) {

	v1obj, ok := obj.(metav1.Object)
	if !ok {
		glog.Infof("not a v1.Object: %v, %s", obj, key)
		return
	}

	annotations := v1obj.GetAnnotations()
	if val, ok := annotations[CloudProivderAnnotation]; ok {
		// match on cloud-provider == oci
		if val == OCIProviderValue {
			c.queue.Add(key)
		}
	} else {
		// for oci-manager lets default to process if annotation doesn't exist
		c.queue.Add(key)
	}
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

	object, err := c.reconcile(key.(string))
	if err == nil {
		glog.Infof("Reconciled %v", key)
		err := c.update(key.(string), object)
		if err != nil {
			c.queue.AddRateLimited(key)
		} else {
			c.queue.Forget(key)
		}
	} else if c.queue.NumRequeues(key) < MAX_REQUEUES {
		glog.Errorf("Reconciled %v with error %v ... retrying", key, err)
		c.queue.AddRateLimited(key)
	} else {
		glog.Errorf("Reconciled %v with error %v - skipping since hit max requeues: %v", key, err, MAX_REQUEUES)
		c.queue.Forget(key)
		utilruntime.HandleError(err)
	}

	return true
}

func (c *Controller) update(key string, object runtime.Object) error {
	kind := c.adapter.Kind()
	if object != nil {
		glog.V(1).Infof("Updating object %s  %s \n", kind, key)
		if _, e := c.adapter.Update(object); e != nil {
			glog.Errorf("ERROR updating object kind %s and key %s: %#v\n", kind, key, e)
			return e
		}
	}
	return nil
}

func (c *Controller) reconcile(key string) (runtime.Object, error) {

	kind := c.adapter.Kind()

	glog.V(2).Infof("Starting to reconcile %v %v\n", kind, key)
	startTime := time.Now()
	defer glog.V(2).Infof("Finished reconciling %v %v (duration: %v)\n", kind, key, time.Now().Sub(startTime))

	obj, exists, err := c.informer.GetStore().GetByKey(key)

	if err != nil {
		return nil, fmt.Errorf("ERROR fetching cloud object with key %s from store: %v", key, err)
	}

	if !exists {
		return nil, nil
	}

	// Copy the object first instead of updating the one in the cache
	source := obj.(runtime.Object)
	object := source.DeepCopyObject()

	objectmeta := c.adapter.ObjectMeta(object)

	if objectmeta.DeletionTimestamp != nil {

		object, err = c.adapter.Delete(object)
		if err != nil {
			glog.Errorf("ERROR deleting cloud object kind %s and key %s: %#v\n", kind, key, err)
			return object, err
		}

		return object, nil

	}

	if objectmeta.DeletionTimestamp == nil && len(objectmeta.GetFinalizers()) == 0 {
		//this is new object we need to set finalizer to proper handle deletes
		//actual processing happens on next cycle
		objectmeta.SetFinalizers([]string{cloudv1alpha1.GroupName})
		return object, nil
	}

	reconciled, err := c.adapter.Reconcile(object)

	if err != nil {
		glog.Errorf("ERROR reconcile cloud object kind %s and key %s: %#v\n", kind, key, err)
	}

	if c.adapter.Equivalent(source, reconciled) {
		glog.V(4).Infof("Match cloud object %s  %s\n", kind, key)
		return nil, nil
	}
	return reconciled, nil
}
