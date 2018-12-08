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
package compute

import (
	"errors"
	"k8s.io/client-go/kubernetes"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	cloudv1alpha1 "github.com/oracle/oci-manager/pkg/apis/cloud.k8s.io/v1alpha1"
	common "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	ocicorev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	lbv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocilb.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	cloudcommon "github.com/oracle/oci-manager/pkg/controller/oci/cloud/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type ComputeAdapter struct {
	subscribtions []schema.GroupVersionResource
	clientset     versioned.Interface
	lister        cache.GenericLister
	queue         workqueue.RateLimitingInterface
}

var controllerKind = cloudv1alpha1.SchemeGroupVersion.WithKind(cloudv1alpha1.ComputeKind)

// init to register compute cloud type
func init() {
	cloudcommon.RegisterCloudType(
		cloudv1alpha1.ComputeResourcePlural,
		cloudv1alpha1.ComputeKind,
		cloudv1alpha1.GroupName,
		&cloudv1alpha1.ComputeValidation,
		NewComputeAdapter,
	)
}

// factory method
func NewComputeAdapter(clientSet versioned.Interface, kubeclient kubernetes.Interface) cloudcommon.CloudTypeAdapter {
	na := ComputeAdapter{
		clientset: clientSet,
	}
	na.subscribtions = subscribe()

	rand.Seed(time.Now().UTC().UnixNano())
	return &na
}

// subscribe to subnet and instance events
func subscribe() []schema.GroupVersionResource {
	subs := make([]schema.GroupVersionResource, 0)
	subs = append(subs, ocicorev1alpha1.SchemeGroupVersion.WithResource(ocicorev1alpha1.SubnetResourcePlural))
	subs = append(subs, ocicorev1alpha1.SchemeGroupVersion.WithResource(ocicorev1alpha1.InstanceResourcePlural))
	return subs
}

// set lister
func (a *ComputeAdapter) SetLister(lister cache.GenericLister) {
	a.lister = lister
}

// set queue
func (a *ComputeAdapter) SetQueue(q workqueue.RateLimitingInterface) {
	a.queue = q
}

// kind
func (a *ComputeAdapter) Kind() string {
	return cloudv1alpha1.ComputeKind
}

// resource
func (a *ComputeAdapter) Resource() string {
	return cloudv1alpha1.ComputeResourcePlural
}

// group version with resource
func (a *ComputeAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return cloudv1alpha1.SchemeGroupVersion.WithResource(cloudv1alpha1.ComputeResourcePlural)
}

// subscriptions
func (a *ComputeAdapter) Subscriptions() []schema.GroupVersionResource {
	return a.subscribtions
}

// object meta
func (a *ComputeAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*cloudv1alpha1.Compute).ObjectMeta
}

// equivalent
func (a *ComputeAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	compute1 := obj1.(*cloudv1alpha1.Compute)
	compute2 := obj2.(*cloudv1alpha1.Compute)

	status_equal := reflect.DeepEqual(compute1.Status, compute2.Status)
	glog.Infof("Equivalent Status: %v", status_equal)
	spec_equal := reflect.DeepEqual(compute1.Spec, compute2.Spec)
	glog.Infof("Equivalent Spec: %v", spec_equal)

	equal := spec_equal && status_equal &&
		compute1.Status.State != cloudv1alpha1.OperatorStatePending

	glog.Infof("Equivalent: %v", equal)

	return equal
}

// delete compute cloud event handler
func (a *ComputeAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	compute := obj.(*cloudv1alpha1.Compute)
	glog.Infof("start compute delete...")
	availabilityDomains, err := cloudcommon.GetAvailabilityDomains(a.clientset, compute.Namespace, compute.Namespace)
	if err != nil {
		return compute, err
	}

	wait := false

	subnetOffset, err := cloudcommon.GetSubnetOffset(a.clientset, compute.Namespace, compute.Spec.Network, compute.Name)
	if err != nil {
		glog.Errorf("error getting subnet offset: %v", err)
		cloudcommon.SetCondition(&compute.Status.OperatorStatus, "GetSubnetOffset", err.Error())
		return compute, err
	}

	// Delete the subnet resource
	for _, availabilityDomain := range availabilityDomains {
		glog.Infof("deleting subnet for ad: %s", availabilityDomain)
		subnetName := compute.Name + "-" + strconv.Itoa(subnetOffset+1)
		subnet, err := DeleteSubnet(a.clientset, compute.Namespace, subnetName)
		if err != nil {
			cloudcommon.SetCondition(&compute.Status.OperatorStatus, v1alpha1.SubnetKind, err.Error())
			return compute, err
		}
		if subnet != nil {
			err = errors.New("subnet resources are not deleted yet")
			cloudcommon.SetCondition(&compute.Status.OperatorStatus, v1alpha1.SubnetKind, err.Error())
			wait = true
		}
	}

	// Delete the instance resources
	for ord := compute.Spec.Replicas - 1; ord > -1; ord-- {
		instance, err := DeleteInstance(a.clientset, compute.Namespace, compute.Name+"-"+strconv.Itoa(ord))
		if err != nil {
			cloudcommon.SetCondition(&compute.Status.OperatorStatus, v1alpha1.InstanceKind, err.Error())
			return compute, err
		}
		if instance != nil {
			err = errors.New("instance resource is not deleted yet")
			cloudcommon.SetCondition(&compute.Status.OperatorStatus, v1alpha1.InstanceKind, err.Error())
			wait = true
		}

	}

	network, err := a.clientset.CloudV1alpha1().Networks(compute.Namespace).Get(compute.Spec.Network, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("could not get network: %s - err: %v", compute.Spec.Network, err)
	}
	delete(network.Status.SubnetAllocationMap, compute.Name)
	_, err = a.clientset.CloudV1alpha1().Networks(compute.Namespace).Update(network)
	if err != nil {
		glog.Errorf("could not update network: %s - err: %v", compute.Spec.Network, err)
	}

	// Remove finalizers
	if !wait && len(compute.Finalizers) > 0 {
		compute.SetFinalizers([]string{})
		return compute, nil
	}

	return compute, nil
}

// update compute cloud event handler
func (a *ComputeAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	compute := obj.(*cloudv1alpha1.Compute)
	glog.Infof("start compute update...")
	resultObj, e := a.clientset.CloudV1alpha1().Computes(compute.Namespace).Update(compute)
	return resultObj, e
}

// reconcile - handles create and updates
func (a *ComputeAdapter) Reconcile(obj runtime.Object) (runtime.Object, error) {
	compute := obj.(*cloudv1alpha1.Compute)

	if compute.Spec.Template.OsType == "" {
		compute.Spec.Template.OsType = "oracle-linux"
	}

	controllerRef := cloudcommon.CreateControllerRef(compute, controllerKind)
	instanceNameMap := make(map[string]string)
	reconcileState := cloudv1alpha1.OperatorStateCreated

	// 1-time get from compartment and randomize
	var availabilityDomains []string
	if len(compute.Status.AvailabilityZones) == 0 {

		ads, err := cloudcommon.GetAvailabilityDomains(a.clientset, compute.Namespace, compute.Namespace)
		if err != nil {
			cloudcommon.SetCondition(&compute.Status.OperatorStatus, "AvailabilityDomain", err.Error())
			return compute, err
		}
		availabilityDomains = make([]string, len(ads))
		perm := rand.Perm(len(ads))
		for i, v := range perm {
			availabilityDomains[v] = ads[i]
		}
		compute.Status.AvailabilityZones = availabilityDomains
		glog.Infof("ad set: %v", availabilityDomains)

	} else {
		availabilityDomains = compute.Status.AvailabilityZones
	}

	adCount := len(availabilityDomains)

	allSubnetsReady := true
	allInstancesReady := true

	instanceReadyCount := 0

	subnetOffset, err := cloudcommon.GetSubnetOffset(a.clientset, compute.Namespace, compute.Spec.Network, compute.Name)
	if err != nil {
		glog.Errorf("error getting subnet offset: %v", err)
		cloudcommon.SetCondition(&compute.Status.OperatorStatus, "GetSubnetOffset", err.Error())
		return compute, nil
	}

	network, err := a.clientset.CloudV1alpha1().Networks(compute.Namespace).Get(compute.Spec.Network, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("error getting network: %v", err)
		cloudcommon.SetCondition(&compute.Status.OperatorStatus, "GetNetwork", err.Error())
		return compute, nil
	}

	if network.Spec.CidrBlock == "" {
		cloudcommon.SetCondition(&compute.Status.OperatorStatus, "GetNetwork", "cidr on network is empty")
		return compute, nil
	}
	networkOctets := strings.Split(network.Spec.CidrBlock, ".")

	subnetMap := make(map[string]string, adCount)
	// Process the subnet and instance resources
	for i, availabilityDomain := range availabilityDomains {
		subnetOctet := subnetOffset + i + 1
		cidrBlock := networkOctets[0] + "." + networkOctets[1] + "." + strconv.Itoa(subnetOctet) + ".0/24"
		subnetName := strconv.Itoa(subnetOctet)
		subnetMap[availabilityDomain] = compute.Name + "-" + subnetName
		subnet, _, err := CreateOrUpdateSubnet(
			a.clientset,
			compute.Namespace,
			cloudv1alpha1.ComputeKind,
			compute.Name,
			compute.Spec.Network,
			controllerRef,
			subnetName,
			availabilityDomain,
			cidrBlock,
			&compute.Spec.SecuritySelector)

		if err != nil {
			cloudcommon.SetCondition(&compute.Status.OperatorStatus, v1alpha1.SubnetKind, err.Error())
			return compute, nil
		}

		if subnet.Status.State != common.ResourceStateProcessed || !subnet.IsResource() {
			glog.Infof("subnet: %s not ready", subnet.Name)
			glog.Infof("compute: %s", compute.Name)
			cloudcommon.SetCondition(&compute.Status.OperatorStatus, v1alpha1.SubnetKind, "subnet not ready")
			reconcileState = cloudv1alpha1.OperatorStatePending
			allSubnetsReady = false
			allInstancesReady = false
		}
	}

	if allSubnetsReady {
		cloudcommon.RemoveCondition(&compute.Status.OperatorStatus, v1alpha1.SubnetKind)
		adIndex := 0
		for ordinal := 0; ordinal < compute.Spec.Replicas; ordinal++ {
			if adIndex >= adCount {
				adIndex = 0
			}
			availabilityDomain := availabilityDomains[adIndex]
			subnetName := subnetMap[availabilityDomain]
			instanceName := getInstanceName(compute, ordinal)
			instanceNameMap[instanceName] = ""
			instance, _, err := CreateOrUpdateInstance(a.clientset, compute, controllerRef, &availabilityDomain, &subnetName, &instanceName)

			if err != nil {
				cloudcommon.SetCondition(&compute.Status.OperatorStatus, v1alpha1.InstanceKind, err.Error())
				return compute, nil
			}

			if instance.Status.State != common.ResourceStateProcessed || !instance.IsResource() {
				err = errors.New("instance resources are not processed yet")
				cloudcommon.SetCondition(&compute.Status.OperatorStatus, v1alpha1.InstanceKind, err.Error())
				reconcileState = cloudv1alpha1.OperatorStatePending
				allInstancesReady = false
			} else {
				instanceReadyCount++
			}
			adIndex++
		}

		// delete orphans - ie when scales down
		selector := "compute=" + compute.Name
		listOptions := metav1.ListOptions{LabelSelector: selector}
		instances, err := a.clientset.OcicoreV1alpha1().Instances(compute.Namespace).List(listOptions)
		if err != nil {
			glog.Errorf("could not list instances: %v", err)
		} else {
			for _, instance := range instances.Items {
				if _, ok := instanceNameMap[instance.Name]; ok {
					// do nothing - should be there
				} else {
					if instance.Status.Dependents != nil {
						for kind, vals := range instance.Status.Dependents {
							if kind == lbv1alpha1.BackendKind {
								for _, val := range vals {
									depParts := strings.Split(val, "/")
									backend := depParts[1]
									glog.Infof("deleting backend: %s for orphan instance: %s", backend, instance.Name)
									err = a.clientset.OcilbV1alpha1().Backends(compute.Namespace).Delete(backend, nil)
									if err != nil {
										glog.Errorf("could not delete orphan backend: %s - err: %v", backend, err)
									}
								}
							}
						}
					}
					err = a.clientset.OcicoreV1alpha1().Instances(compute.Namespace).Delete(instance.Name, nil)
					if err != nil {
						glog.Errorf("could not delete orphan instance: %s - err: %v", instance.Name, err)
					}
				}
			}
		}

		compute.Status.ReadyReplicas = instanceReadyCount
		compute.Status.UnavailableReplicas = compute.Spec.Replicas - instanceReadyCount
	}

	if allInstancesReady {
		cloudcommon.RemoveCondition(&compute.Status.OperatorStatus, v1alpha1.InstanceKind)
	}

	// Everything is done. Update the State, reset the Conditions and return
	if reconcileState == cloudv1alpha1.OperatorStateCreated {
		compute.Status.State = cloudv1alpha1.OperatorStateCreated
		if compute.Status.Conditions != nil {
			compute.Status.Conditions = []cloudv1alpha1.OperatorCondition{}
		}
	} else {
		compute.Status.State = cloudv1alpha1.OperatorStatePending
	}
	glog.Infof("thru reconcile")
	return compute, nil

}

// callback for resource
func (a *ComputeAdapter) CallbackForResource(resource schema.GroupVersionResource) cache.ResourceEventHandlerFuncs {
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
func (a *ComputeAdapter) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *cloudv1alpha1.Compute {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}
	obj, err := a.lister.ByNamespace(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}

	compute := obj.(*cloudv1alpha1.Compute)

	if compute.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return compute
}

// ignore add event
func (a *ComputeAdapter) ignoreAddEvent(obj interface{}) {
	glog.V(5).Infof("Got add event: %v, ignore", obj)
}

// handle update or delete events
func (a *ComputeAdapter) processUpdateOrDeleteEvent(obj interface{}) {
	object := obj.(metav1.Object)
	if controllerRef := cloudcommon.GetControllerOf(object); controllerRef != nil {
		compute := a.resolveControllerRef(object.GetNamespace(), controllerRef)
		if compute == nil {
			return
		}
		key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(compute)
		if err != nil {
			glog.Errorf("Compute deletion state error %v", err)
			return
		}
		glog.V(4).Infof("Compute %s received update event for %s %s\n", key, reflect.TypeOf(object).String(), object.GetName())
		a.queue.Add(key)
		//a.Reconcile(compute)
		return
	}
}
