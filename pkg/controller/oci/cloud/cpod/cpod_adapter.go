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
package cpod

import (
	"errors"
	"github.com/golang/glog"
	cloudv1alpha1 "github.com/oracle/oci-manager/pkg/apis/cloud.k8s.io/v1alpha1"
	ocicommon "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	ocicorev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	cloudcommon "github.com/oracle/oci-manager/pkg/controller/oci/cloud/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"reflect"
)

type CpodAdapter struct {
	subscribtions []schema.GroupVersionResource
	clientset     versioned.Interface
	lister        cache.GenericLister
	queue         workqueue.RateLimitingInterface
}

const (
	DefaultShape                 string = "VM.Standard1.1"
	DefaultCpodInstanceOsType    string = "Oracle-Linux"
	DefaultCpodInstanceOSVersion string = "7.5-2018.05.09-1"
	DefaultCpodHostPort          string = "80"
)

var controllerKind = cloudv1alpha1.SchemeGroupVersion.WithKind(cloudv1alpha1.CpodKind)

// init to register cpod cloud type
func init() {
	cloudcommon.RegisterCloudType(
		cloudv1alpha1.CpodResourcePlural,
		cloudv1alpha1.CpodKind,
		cloudv1alpha1.GroupName,
		nil,
		NewCpodAdapter,
	)
}

// factory method
func NewCpodAdapter(clientSet versioned.Interface, kubeclient kubernetes.Interface) cloudcommon.CloudTypeAdapter {
	na := CpodAdapter{
		clientset: clientSet,
	}
	na.subscribtions = subscribe()
	return &na
}

// subscribe to subnet and instance events
func subscribe() []schema.GroupVersionResource {
	subs := make([]schema.GroupVersionResource, 0)
	subs = append(subs, ocicorev1alpha1.SchemeGroupVersion.WithResource(ocicorev1alpha1.InstanceResourcePlural))
	return subs
}

// set lister
func (a *CpodAdapter) SetLister(lister cache.GenericLister) {
	a.lister = lister
}

// set queue
func (a *CpodAdapter) SetQueue(q workqueue.RateLimitingInterface) {
	a.queue = q
}

// kind
func (a *CpodAdapter) Kind() string {
	return cloudv1alpha1.CpodKind
}

// resource
func (a *CpodAdapter) Resource() string {
	return cloudv1alpha1.CpodResourcePlural
}

// group version with resource
func (a *CpodAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return cloudv1alpha1.SchemeGroupVersion.WithResource(cloudv1alpha1.CpodResourcePlural)
}

// subscriptions
func (a *CpodAdapter) Subscriptions() []schema.GroupVersionResource {
	return a.subscribtions
}

// object meta
func (a *CpodAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*cloudv1alpha1.Cpod).ObjectMeta
}

// equivalent
func (a *CpodAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	return reflect.DeepEqual(obj1, obj2)
}

// delete cpod cloud event handler
func (a *CpodAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	cpod := obj.(*cloudv1alpha1.Cpod)

	wait := false

	// Delete the cpod instance resource
	cpodInstance, err := DeleteCpodInstance(a.clientset, cpod)
	if err != nil {
		cloudcommon.SetCondition(&cpod.Status.OperatorStatus, ocicorev1alpha1.InstanceKind, err.Error())
		return cpod, err
	}
	if cpodInstance != nil {
		err = errors.New("Cpod instance resource is not deleted yet")
		cloudcommon.SetCondition(&cpod.Status.OperatorStatus, ocicorev1alpha1.InstanceKind, err.Error())
		wait = true
	}

	// Remove finalizers
	if !wait && len(cpod.Finalizers) > 0 {
		cpod.SetFinalizers([]string{})
		return cpod, nil
	}

	return cpod, nil
}

// update cpod cloud event handler
func (a *CpodAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	cpod := obj.(*cloudv1alpha1.Cpod)
	resultObj, e := a.clientset.CloudV1alpha1().Cpods(cpod.Namespace).Update(cpod)
	return resultObj, e
}

// reconcile - handles create and updates
func (a *CpodAdapter) Reconcile(obj runtime.Object) (runtime.Object, error) {
	cpod := obj.(*cloudv1alpha1.Cpod)
	controllerRef := cloudcommon.CreateControllerRef(cpod, controllerKind)

	reconcileState := cloudv1alpha1.OperatorStateCreated

	instance, _, err := CreateOrUpdateCpodInstance(a.clientset, cpod, controllerRef)

	if err != nil {
		cloudcommon.SetCondition(&cpod.Status.OperatorStatus, ocicorev1alpha1.InstanceKind, err.Error())
		return cpod, err
	}

	if instance.Status.State != ocicommon.ResourceStateProcessed || !instance.IsResource() {
		err = errors.New("instance resources are not processed yet")
		cloudcommon.SetCondition(&cpod.Status.OperatorStatus, ocicorev1alpha1.InstanceKind, err.Error())
		reconcileState = cloudv1alpha1.OperatorStatePending
	}

	/*
		if reconcileState == cloudv1alpha1.OperatorStateCreated {
			cpod.Status.State = cloudv1alpha1.OperatorStateCreated
			cpod.Status.Conditions = []cloudv1alpha1.OperatorCondition{}
			cpod.Status.PodStatus.Phase = corev1.PodRunning
			cpod.Status.PodStatus.PodIP = *instance.Status.PrimaryVnic.PublicIp
			if cpod.Status.PodStatus.HostIP == "" {
				cpod.Status.PodStatus.HostIP = *instance.Status.PrimaryVnic.PublicIp
			}
		} else {
			cpod.Status.State = cloudv1alpha1.OperatorStatePending
			cpod.Status.PodStatus.Phase = corev1.PodPending
		}
	*/

	// Everything is done. Update the State, reset the Conditions and return
	podStatus := a.getPodState(reconcileState, instance)
	if reconcileState == cloudv1alpha1.OperatorStateCreated {
		if cpod.Status.PodStatus.HostIP == "" {
			podStatus.HostIP = *instance.Status.PrimaryVnic.PublicIp
		} else {
			podStatus.HostIP = cpod.Status.PodStatus.HostIP
		}
		podStatus.PodIP = *instance.Status.PrimaryVnic.PublicIp
		cpod.Status.Conditions = []cloudv1alpha1.OperatorCondition{}
	}

	cpod.Status.PodStatus = *podStatus
	cpod.Status.State = reconcileState
	return cpod, nil

}

func (a *CpodAdapter) getPodState(reconcileState cloudv1alpha1.OperatorState, inst *ocicorev1alpha1.Instance) *corev1.PodStatus {
	podName := inst.Name

	startedAt := inst.CreationTimestamp
	var (
		//podCondition   corev1.PodCondition
		containerState corev1.ContainerState
		podPhase       corev1.PodPhase
		conditions     []corev1.PodCondition
	)
	conditions = make([]corev1.PodCondition, 0)
	switch reconcileState {
	case cloudv1alpha1.OperatorStatePending:
		podCondition := corev1.PodCondition{
			Type:   corev1.PodScheduled,
			Status: corev1.ConditionFalse,
		}
		conditions = append(conditions, podCondition)
		podPhase = corev1.PodPending
		containerState = corev1.ContainerState{
			Waiting: &corev1.ContainerStateWaiting{},
		}
	case cloudv1alpha1.OperatorStateCreated: // running
		podConditionReady := corev1.PodCondition{
			Type:   corev1.PodReady,
			Status: corev1.ConditionTrue,
		}
		conditions = append(conditions, podConditionReady)

		podConditionScheduled := corev1.PodCondition{
			Type:   corev1.PodScheduled,
			Status: corev1.ConditionTrue,
		}
		conditions = append(conditions, podConditionScheduled)

		podConditionInited := corev1.PodCondition{
			Type:   corev1.PodInitialized,
			Status: corev1.ConditionTrue,
		}
		conditions = append(conditions, podConditionInited)

		podPhase = corev1.PodRunning
		containerState = corev1.ContainerState{
			Running: &corev1.ContainerStateRunning{
				StartedAt: startedAt,
			},
		}
	default: //unkown
		podCondition := corev1.PodCondition{
			Type:   corev1.PodReasonUnschedulable,
			Status: corev1.ConditionUnknown,
		}
		conditions = append(conditions, podCondition)
		podPhase = corev1.PodUnknown
		containerState = corev1.ContainerState{}
	}

	status := corev1.PodStatus{
		Phase:      podPhase,
		Conditions: conditions,
		Message:    inst.Status.Message,
		Reason:     "",
		ContainerStatuses: []corev1.ContainerStatus{
			{
				Name:         podName,
				RestartCount: int32(0),
				Image:        "fake",
				ImageID:      "fake",
				ContainerID:  "fake",
				Ready:        true,
				State:        containerState,
			},
		},
	}
	return &status

}

// callback for resource
func (a *CpodAdapter) CallbackForResource(resource schema.GroupVersionResource) cache.ResourceEventHandlerFuncs {
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			a.ignoreAddEvent(obj)
		},
		UpdateFunc: func(old, obj interface{}) {
			a.processUpdateOrDeleteEvent(obj)
		},
		DeleteFunc: func(obj interface{}) {
			a.processUpdateOrDeleteEvent(obj)
		},
	}

	return handlers
}

// resolve controller reference
func (a *CpodAdapter) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *cloudv1alpha1.Cpod {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}
	obj, err := a.lister.ByNamespace(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}

	cpod := obj.(*cloudv1alpha1.Cpod)

	if cpod.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return cpod
}

// ignore add event
func (a *CpodAdapter) ignoreAddEvent(obj interface{}) {
	glog.V(5).Infof("Got add event: %v, ignore", obj)
}

// handle update or delete events
func (a *CpodAdapter) processUpdateOrDeleteEvent(obj interface{}) {
	object := obj.(metav1.Object)
	if controllerRef := cloudcommon.GetControllerOf(object); controllerRef != nil {
		cpod := a.resolveControllerRef(object.GetNamespace(), controllerRef)
		if cpod == nil {
			return
		}
		key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(cpod)
		if err != nil {
			glog.Errorf("Cpod deletion state error %v", err)
			return
		}
		glog.V(4).Infof("Cpod %s received update event for %s %s\n", key, reflect.TypeOf(object).String(), object.GetName())
		a.queue.Add(key)
		//a.Reconcile(cpod)
		return
	}
}
