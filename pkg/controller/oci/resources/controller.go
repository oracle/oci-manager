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

package resources

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
	ocisdkcommon "github.com/oracle/oci-go-sdk/common"
	ocicommon "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	clientset "github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	clientsetScheme "github.com/oracle/oci-manager/pkg/client/clientset/versioned/scheme"
	informers "github.com/oracle/oci-manager/pkg/client/informers/externalversions"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
)

// OciGroupName constant is used for finalizers string
const (
	OciGroupName            = "oci.oracle.com"
	eventTypeResourceUpdate = "ResourceUpdate"
	eventTypeResourceError  = "ResourceError"
)

// Controller of resource create/update/delete events
type Controller struct {
	//resourceclient *oci.Client
	queue    workqueue.RateLimitingInterface
	queueMap map[string]workqueue.RateLimitingInterface
	informer cache.SharedInformer
	adapter  resourcescommon.ResourceTypeAdapter
	factory  informers.SharedInformerFactory
	recorder record.EventRecorder
}

// Start a new controller for a type adapter
func Start(
	clientset clientset.Interface,
	kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider,
	informerFactory informers.SharedInformerFactory,
	stopChan <-chan struct{},
	adapterFactory resourcescommon.AdapterFactory,
	adapterSpecificArgs map[string]interface{},
	queueMap map[string]workqueue.RateLimitingInterface,
) *Controller {
	adapter := adapterFactory(clientset, kubeclient, ociconfig, adapterSpecificArgs)
	controller := New(adapter, kubeclient, informerFactory, queueMap)
	controller.Run(stopChan)
	return controller
}

// New initializes a controller object
func New(adapter resourcescommon.ResourceTypeAdapter,
	kubeclient kubernetes.Interface,
	informerFactory informers.SharedInformerFactory,
	queueMap map[string]workqueue.RateLimitingInterface) *Controller {
	c := &Controller{
		adapter:  adapter,
		queue:    queueMap[adapter.Kind()],
		queueMap: queueMap,
		factory:  informerFactory,
	}

	glog.V(4).Infof("Creating event broadcaster for resource %s", adapter.Kind())
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(clientsetScheme.Scheme, corev1.EventSource{Component: adapter.Kind()})

	c.recorder = recorder

	glog.V(4).Infof("Building informer for resource - %v", adapter.Resource())

	genericInfomer, err := informerFactory.ForResource(adapter.GroupVersionWithResource())

	if err != nil {
		glog.Fatalf("Error building informer for resource: %s - %v", adapter.Resource(), err)
	}

	c.informer = genericInfomer.Informer()

	c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				// fmt.Printf("EVENT ADD key %s: %#v\n", key, obj)
				c.queue.Add(key)
			}
		},
		UpdateFunc: func(old, cur interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(cur)
			if err == nil {
				// fmt.Printf("EVENT UPDATE %v %v\n", old, cur)
				c.queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				// fmt.Printf("EVENT DELETE %v\n", obj)
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
		if retry && c.queue.NumRequeues(key) <= 50 {
			glog.V(4).Infof("Resubmit for reconcile key: %v", key)
			c.queue.AddAfter(key, 3*time.Second)
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
		if apierrors.IsConflict(err) {
			glog.V(4).Infof("Conflict during reconcile %s key %v - %v", c.adapter.Kind(), key, err)
		} else {
			glog.Errorf("Error reconciling %s key %v - %v", c.adapter.Kind(), key, err)
		}
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
	kind := c.adapter.Kind()
	if object != nil {
		glog.V(1).Infof("Updating object %s  %s \n", kind, key)
		glog.V(4).Infof("Updating object %s  %s --- %#v\n", kind, key, object)
		if _, e := c.adapter.UpdateObject(object); e != nil {
			if apierrors.IsConflict(e) {
				glog.V(4).Infof("Conflict updating reconciled CRD %s key %v - %v", c.adapter.Kind(), key, e)
			} else {
				glog.Errorf("ERROR updating object kind %s and key %s: %#v\n", kind, key, e)
			}
			return e
		}
		c.recorder.Event(object, corev1.EventTypeNormal, eventTypeResourceUpdate, fmt.Sprintf("Updated CRD resource %s  %s", kind, key))

		//signal all Dependents about parent update
		for depKind, depKeys := range c.adapter.Dependents(object) {
			for _, depKey := range depKeys {
				if _, ok := c.queueMap[depKind]; ok {
					glog.V(4).Infof("Signal to %s, %s, about update on %s %#v", depKind, depKey, key, object)
					c.queueMap[depKind].Add(depKey)
					c.recorder.Event(object, corev1.EventTypeNormal, eventTypeResourceUpdate, fmt.Sprintf("Signal to %s, %s, about update", depKind, depKey))

				}
			}
		}
	}
	return nil
}

// Reconcile the resource with the incoming object specification.
// This is the main func doing the work to manage the lifecycle of a resource
func (c *Controller) reconcile(key string) (reconciled runtime.Object, err error, retry bool) {

	kind := c.adapter.Kind()

	glog.V(3).Infof("Starting to reconcile %v %v\n", kind, key)
	startTime := time.Now()
	defer glog.V(3).Infof("Finished reconciling %v %v (duration: %v)\n", kind, key, time.Now().Sub(startTime))

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

	if objectmeta.DeletionTimestamp == nil && len(objectmeta.GetFinalizers()) == 0 {
		//this is new object we need to set finalizer to proper handle deletes
		//actual processing happens on next cycle
		objectmeta.SetFinalizers([]string{OciGroupName})
		return object, nil, false
	}

	// Delete
	// If we have a pending delete first to start the delete flow
	// by removing the remote resource object first and them cleaning
	// up the finalizers from the object and finally completing the delete
	// of the object itself. This will require several reconcile passes to complete
	if objectmeta.DeletionTimestamp != nil {

		if c.haveDeps(object) {
			format := "Dependents still present on resource %s  %s, re-submit for reconcile with backoff \n"
			glog.V(2).Infof(format, kind, key)
			return nil, nil, true
		}

		if c.adapter.Id(object) != "" {

			glog.V(1).Infof("Deleting resource %s  %s \n", kind, key)
			glog.V(4).Infof("Deleting resource %s  %s %#v\n", kind, key, object)
			object, err = c.adapter.Delete(object)
			if err != nil {
				glog.Errorf("ERROR deleting resource kind %s and key %s: %#v\n", kind, key, err)
				return object, err, false
			} else if getResourceState(object) == ocicommon.ResourceStatePending {
				return object, nil, true
			}
		}

		err = c.removeDependencyFromParent(object)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return object, err, false
			}
		}

		if len(objectmeta.GetFinalizers()) > 0 {
			objectmeta.SetFinalizers([]string{})
		}

		return object, nil, false

	}

	//Create
	//check dependencies
	if ready, err := c.isDependencyReady(object); !ready {
		if err != nil {
			return nil, err, false
		}
		return nil, nil, false
	}

	// Create
	// The Id indicates if the resource is already created or it's pending.
	// After setting the finalizer we return to start the reconcile processing
	// from the beginning since the object was updated. It should be skipped on
	// the next loop and proceed with the Get below to validate the create
	if c.adapter.Id(object) == "" {
		glog.V(1).Infof("Creating resource %s  %s \n", kind, key)
		glog.V(5).Infof("Creating resource %s  %s --- %#v\n", kind, key, object)
		created, err := c.adapter.Create(object)
		if err != nil {
			errMsg := fmt.Sprintf("ERROR creating resource kind %s and key %s: %#v\n", kind, key, err)
			glog.Errorf(errMsg)
			c.recorder.Event(object, corev1.EventTypeWarning, eventTypeResourceError, errMsg)
			return object, err, false
		}
		if created != nil {
			c.recorder.Event(created, corev1.EventTypeNormal, eventTypeResourceUpdate, fmt.Sprintf("Created OCI resource %s  %s", kind, key))
		}
		return created, nil, false
	}

	// Get
	// If the current state of the object is not pending a delete or create
	// we proceed with a Get call to fetch the remote resource so we can compare
	// to the current object
	found, err := c.adapter.Get(object)
	if err != nil {
		errMsg := fmt.Sprintf("ERROR getting resource kind %s and key %s: %#v\n", kind, key, err)
		glog.Errorf(errMsg)
		c.recorder.Event(object, corev1.EventTypeWarning, eventTypeResourceError, errMsg)
		return found, err, false
	} else if getResourceState(found) == ocicommon.ResourceStatePending {
		return found, nil, true
	}

	// Update
	// If the object we got back from the Get call above differs from the incoming
	// object from queue we proceed with the resource update. The delta can be due
	// to a change in the spec of the object that needs to be reconciled with the
	// current resource or it could be because of some external change of the resource
	// that needs to be corrected since we are the source of truth for the resource

	if found != nil {
		if !c.adapter.IsResourceCompliant(found) {
			glog.V(1).Infof("Updating resource %s  %s %#v\n", kind, key, found)
			//this updates underlying oci resource not the crd
			// crd will get updated in report func
			updated, err := c.adapter.Update(object)
			if err != nil {
				errMsg := fmt.Sprintf("ERROR updating resource kind %s and key %s: %#v\n", kind, key, err)
				glog.Errorf(errMsg)
				c.recorder.Event(object, corev1.EventTypeWarning, eventTypeResourceError, errMsg)
				return object, err, true
			}
			if updated != nil {
				c.recorder.Event(updated, corev1.EventTypeNormal, eventTypeResourceUpdate, fmt.Sprintf("Updated OCI resource %s  %s", kind, key))
			}

			return updated, nil, false
		}

		if c.adapter.IsResourceStatusChanged(source, found) {
			//TODO add adapter call to get current resource status
			logMsg := fmt.Sprintf("LifeCycleState of resource %s changed", key)
			glog.V(1).Info(logMsg)
			c.recorder.Event(object, corev1.EventTypeNormal, eventTypeResourceUpdate, logMsg)
			return found, nil, false
		}
	}

	// Finally if there is no change it's a no-op
	glog.V(3).Infof("Match resource %s  %s\n", kind, key)
	return nil, nil, false

}

func (c *Controller) isDependencyReady(obj runtime.Object) (bool, error) {

	glog.V(4).Infof("Checking parent dependency on %s %#v", obj.GetObjectKind().GroupVersionKind().Kind, obj)

	var deps []runtime.Object

	labelSelectorsMap := c.adapter.DependsOn(obj)

	if labelSelectorsMap != nil && len(labelSelectorsMap) > 0 {
		for kinds, selector := range labelSelectorsMap {
			sgvr := c.adapter.GroupVersionWithResource()
			sgvr.Resource = kinds
			informer, err := c.factory.ForResource(sgvr)
			if err != nil {
				return false, err
			}
			selector := labels.Set(selector.LabelSelector).AsSelectorPreValidated()
			depOns, err := informer.Lister().List(selector)
			if err != nil || len(depOns) == 0 {
				glog.V(4).Infof("No ready parents found for %s %#v", obj.GetObjectKind().GroupVersionKind().Kind, obj)
				return false, err
			}
			deps = append(deps, depOns...)
		}
	}

	depRefs, err := c.adapter.DependsOnRefs(obj)
	if err != nil {
		return false, err
	}

	deps = append(deps, depRefs...)

	for _, dep := range deps {
		depCopy := dep.DeepCopyObject()
		depObj := depCopy.(ocicommon.ObjectInterface)
		if reged, err := depObj.IsDependentRegistered(c.adapter.Kind(), obj); !reged && err == nil {
			depObj.AddDependent(c.adapter.Kind(), obj)

			_, err = c.adapter.UpdateForResource(depObj.GetGroupVersionResource(), depCopy)

			if err != nil {
				return false, err
			}
		}
		if depObj.GetResourceID() == "" {
			glog.V(4).Infof("Parent %#v is not ready", dep)
			return false, nil
		}
	}

	return true, nil
}

func (c *Controller) removeDependencyFromParent(obj runtime.Object) error {

	glog.V(4).Infof("Removing parent dependency of %s %#v", obj.GetObjectKind().GroupVersionKind().Kind, obj)

	var deps []runtime.Object

	labelSelectorsMap := c.adapter.DependsOn(obj)

	if labelSelectorsMap != nil && len(labelSelectorsMap) > 0 {
		for kinds, selector := range labelSelectorsMap {
			glog.V(2).Infof("Removing dependency from %s", kinds)
			sgvr := c.adapter.GroupVersionWithResource()
			sgvr.Resource = kinds
			informer, err := c.factory.ForResource(sgvr)
			if err != nil {
				return err
			}
			selector := labels.Set(selector.LabelSelector).AsSelectorPreValidated()
			depOns, err := informer.Lister().List(selector)
			if err != nil {
				return err
			}
			deps = append(deps, depOns...)
		}
	}

	depRefs, err := c.adapter.DependsOnRefs(obj)
	if err != nil {
		return err
	}

	deps = append(deps, depRefs...)

	depsFound := false
	for _, dep := range deps {
		depCopy := dep.DeepCopyObject()
		depObj := depCopy.(ocicommon.ObjectInterface)
		if reged, err := depObj.IsDependentRegistered(c.adapter.Kind(), obj); reged && err == nil {
			glog.V(2).Infof("Removing parent dependency of %s/%s from %#v", obj.GetObjectKind().GroupVersionKind().Kind, c.adapter.ObjectMeta(obj).Name, depCopy)
			depObj.RemoveDependent(c.adapter.Kind(), obj)
			_, err = c.adapter.UpdateForResource(depObj.GetGroupVersionResource(), depCopy)
			if err != nil {
				return err
			}
			depsFound = true
		}
	}

	if depsFound {
		return fmt.Errorf("Still have parent dependencies of %s %s", c.adapter.Kind(), c.adapter.ObjectMeta(obj).Name)
	}
	return nil
}

func (c *Controller) haveDeps(obj runtime.Object) bool {
	return c.adapter.Dependents(obj) != nil && len(c.adapter.Dependents(obj)) > 0
}
func getResourceState(obj runtime.Object) ocicommon.ResourceState {
	resource := obj.(ocicommon.ObjectInterface)
	return resource.GetResourceState()
}
