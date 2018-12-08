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
package network

import (
	"errors"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"strconv"
	"strings"

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	cloudv1alpha1 "github.com/oracle/oci-manager/pkg/apis/cloud.k8s.io/v1alpha1"
	common "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	ocicorev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	cloudcommon "github.com/oracle/oci-manager/pkg/controller/oci/cloud/common"
)

func init() {
	cloudcommon.RegisterCloudType(
		cloudv1alpha1.NetworkResourcePlural,
		cloudv1alpha1.NetworkKind,
		cloudv1alpha1.GroupName,
		&cloudv1alpha1.NetworkValidation,
		NewNetworkAdapter,
	)
}

var controllerKind = cloudv1alpha1.SchemeGroupVersion.WithKind("Network")

type NetworkAdapter struct {
	subscribtions []schema.GroupVersionResource
	clientset     versioned.Interface
	lister        cache.GenericLister
	queue         workqueue.RateLimitingInterface
}

func subscribe() []schema.GroupVersionResource {
	subs := make([]schema.GroupVersionResource, 0)
	subs = append(subs, ocicorev1alpha1.SchemeGroupVersion.WithResource(ocicorev1alpha1.VirtualNetworkResourcePlural))
	subs = append(subs, ocicorev1alpha1.SchemeGroupVersion.WithResource(ocicorev1alpha1.RouteTableResourcePlural))
	subs = append(subs, ocicorev1alpha1.SchemeGroupVersion.WithResource(ocicorev1alpha1.InternetGatewayResourcePlural))
	// for out of band/cloud compute adapter subnet creations
	// subs = append(subs, ocicorev1alpha1.SchemeGroupVersion.WithResource(ocicorev1alpha1.SubnetResourcePlural))

	return subs
}

func NewNetworkAdapter(clientSet versioned.Interface, kubeclient kubernetes.Interface) cloudcommon.CloudTypeAdapter {
	na := NetworkAdapter{
		clientset: clientSet,
	}

	na.subscribtions = subscribe()

	return &na
}

func (a *NetworkAdapter) SetLister(lister cache.GenericLister) {
	a.lister = lister
}

func (a *NetworkAdapter) SetQueue(q workqueue.RateLimitingInterface) {
	a.queue = q
}

func (a *NetworkAdapter) Kind() string {
	return cloudv1alpha1.NetworkKind
}

func (a *NetworkAdapter) Resource() string {
	return cloudv1alpha1.NetworkResourcePlural
}

func (a *NetworkAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return cloudv1alpha1.SchemeGroupVersion.WithResource(cloudv1alpha1.NetworkResourcePlural)
}

func (a *NetworkAdapter) Subscriptions() []schema.GroupVersionResource {
	return a.subscribtions
}

func (a *NetworkAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*cloudv1alpha1.Network).ObjectMeta
}

func (a *NetworkAdapter) Equivalent(obj1, obj2 runtime.Object) bool {

	net1 := obj1.(*cloudv1alpha1.Network)
	net2 := obj2.(*cloudv1alpha1.Network)

	status_equal := reflect.DeepEqual(net1.Status, net2.Status)
	glog.Infof("Equivalent Status: %v", status_equal)

	spec_equal := reflect.DeepEqual(net1.Spec, net2.Spec)
	glog.Infof("Equivalent Spec: %v", spec_equal)

	equal := spec_equal && status_equal &&
		net1.Status.State != cloudv1alpha1.OperatorStatePending

	glog.Infof("Equivalent: %v", equal)

	return equal
}

func (a *NetworkAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	network := obj.(*cloudv1alpha1.Network)

	wait := false

	// Delete the route table resource
	routetable, err := DeleteRouteTable(a.clientset, network)
	if err != nil {
		cloudcommon.SetCondition(&network.Status.OperatorStatus, v1alpha1.RouteTableKind, err.Error())
		return network, err
	}

	if routetable != nil {
		err = errors.New("route table resource is not deleted yet")
		cloudcommon.SetCondition(&network.Status.OperatorStatus, v1alpha1.RouteTableKind, err.Error())
		wait = true
	}

	// Delete the internet gateway resource
	internetgateway, err := DeleteInternetGateway(a.clientset, network)
	if err != nil {
		cloudcommon.SetCondition(&network.Status.OperatorStatus, v1alpha1.InternetGatewayKind, err.Error())
		return network, err
	}
	if internetgateway != nil {
		err = errors.New("internet gateway resource is not deleted yet")
		cloudcommon.SetCondition(&network.Status.OperatorStatus, v1alpha1.InternetGatewayKind, err.Error())
		wait = true
	}

	// Delete the virtual network resource
	vcn, err := DeleteVcn(a.clientset, network)
	if err != nil {
		cloudcommon.SetCondition(&network.Status.OperatorStatus, v1alpha1.VirtualNetworkKind, err.Error())
		return network, err
	}
	if vcn != nil {
		err = errors.New("virtual network resource is not deleted yet")
		cloudcommon.SetCondition(&network.Status.OperatorStatus, v1alpha1.VirtualNetworkKind, err.Error())
		wait = true
	}

	// Remove finalizers
	if !wait && len(network.Finalizers) > 0 {
		network.SetFinalizers([]string{})
		return network, nil
	}

	return network, nil
}

func (a *NetworkAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	network := obj.(*cloudv1alpha1.Network)
	resultObj, e := a.clientset.CloudV1alpha1().Networks(network.Namespace).Update(network)
	return resultObj, e
}

func (a *NetworkAdapter) Reconcile(obj runtime.Object) (runtime.Object, error) {

	network := obj.(*cloudv1alpha1.Network)

	// default to 10.0/16 if not specified
	if network.Spec.CidrBlock == "" {
		network.Spec.CidrBlock = "10.0.0.0/16"
	}

	controllerRef := cloudcommon.CreateControllerRef(network, controllerKind)
	reconcileState := cloudv1alpha1.OperatorStateCreated

	// Process the virtual network resource
	vcn, _, err := CreateOrUpdateVcn(a.clientset, network, controllerRef)
	if err != nil {
		cloudcommon.SetCondition(&network.Status.OperatorStatus, v1alpha1.VirtualNetworkKind, err.Error())
		return network, err
	}
	if vcn.Status.State != common.ResourceStateProcessed || !vcn.IsResource() {
		err = errors.New("virtual network resource is not processed yet")
		cloudcommon.SetCondition(&network.Status.OperatorStatus, v1alpha1.VirtualNetworkKind, err.Error())
		reconcileState = cloudv1alpha1.OperatorStatePending
	} else {
		cloudcommon.RemoveCondition(&network.Status.OperatorStatus, v1alpha1.VirtualNetworkKind)
	}

	// Process the internet gateway resource
	internetgateway, _, err := CreateOrUpdateInternetGateway(a.clientset, network, controllerRef)
	if err != nil {
		cloudcommon.SetCondition(&network.Status.OperatorStatus, v1alpha1.InternetGatewayKind, err.Error())
		return network, err
	}
	if internetgateway.Status.State != common.ResourceStateProcessed || !internetgateway.IsResource() {
		err = errors.New("internet gateway resource is not processed yet")
		cloudcommon.SetCondition(&network.Status.OperatorStatus, v1alpha1.InternetGatewayKind, err.Error())
		reconcileState = cloudv1alpha1.OperatorStatePending
	} else {
		cloudcommon.RemoveCondition(&network.Status.OperatorStatus, v1alpha1.InternetGatewayKind)
	}

	// Process the route table resource
	routetable, _, err := CreateOrUpdateRouteTable(a.clientset, network, controllerRef)
	if err != nil {
		cloudcommon.SetCondition(&network.Status.OperatorStatus, v1alpha1.RouteTableKind, err.Error())
		return network, err
	}
	if routetable.Status.State != common.ResourceStateProcessed || !routetable.IsResource() {
		err = errors.New("route table resource is not processed yet")
		cloudcommon.SetCondition(&network.Status.OperatorStatus, v1alpha1.RouteTableKind, err.Error())
		reconcileState = cloudv1alpha1.OperatorStatePending
	} else {
		cloudcommon.RemoveCondition(&network.Status.OperatorStatus, v1alpha1.RouteTableKind)
	}

	// reconcile subnet allocation map
	allSubnetsInNs, err := a.clientset.OcicoreV1alpha1().Subnets(network.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if network.Status.SubnetAllocationMap == nil {
		network.Status.SubnetAllocationMap = make(map[string]int)
	}

	currentRangeMap := make(map[int]string)
	for k, v := range network.Status.SubnetAllocationMap {
		currentRangeMap[v] = k
	}
	subnetRangeMap := make(map[int]v1alpha1.Subnet)
	for _, sn := range allSubnetsInNs.Items {
		if sn.Spec.VcnRef == network.Name {
			subnetOctets := strings.Split(sn.Spec.CidrBlock, ".")
			subnetBlockPart, err := strconv.Atoi(subnetOctets[2])

			// floor to tens due to int
			subnetBlock := (subnetBlockPart / 10) * 10
			if err != nil {
				glog.Errorf("error parsing int from subnet: %s octet: %v", sn.Name, subnetOctets[2])
			}
			subnetRangeMap[subnetBlock] = sn

			_, present := currentRangeMap[subnetBlock]
			if !present {
				key := sn.Name
				if val, ok := sn.Labels["LoadBalancer"]; ok {
					key = val
				} else if val, ok := sn.Labels["Compute"]; ok {
					key = val
				}
				network.Status.SubnetAllocationMap[key] = subnetBlock
			}
		}
	}
	// delete orphan from map
	for k, v := range currentRangeMap {
		_, present := subnetRangeMap[k]
		if !present {
			glog.Infof("deleting orphan range: %v from network: %s", k, network.Name)
			delete(network.Status.SubnetAllocationMap, v)
		}
	}

	// Everything is done. Update the State, reset the Conditions and return
	if reconcileState == cloudv1alpha1.OperatorStateCreated {
		network.Status.State = cloudv1alpha1.OperatorStateCreated
		if network.Status.Conditions != nil {
			network.Status.Conditions = []cloudv1alpha1.OperatorCondition{}
		}
	} else {
		network.Status.State = cloudv1alpha1.OperatorStatePending
	}
	return network, nil

}

func (a *NetworkAdapter) CallbackForResource(resource schema.GroupVersionResource) cache.ResourceEventHandlerFuncs {
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if reflect.TypeOf(obj).String() == "*v1alpha1.Subnet" {
				subnet := obj.(*v1alpha1.Subnet)
				glog.Infof("subnet subscription callback for %s - queueing reconcile of network: %s", subnet.Name, subnet.Spec.VcnRef)
				a.queue.Add(subnet.Spec.VcnRef)
			} else {
				a.ignoreAddEvent(obj)
			}
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

func (a *NetworkAdapter) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *cloudv1alpha1.Network {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}
	obj, err := a.lister.ByNamespace(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}

	network := obj.(*cloudv1alpha1.Network)

	if network.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return network
}

func (a *NetworkAdapter) ignoreAddEvent(obj interface{}) {
	glog.V(4).Infof("Got add event: %v - type: %s", obj, reflect.TypeOf(obj))
}

func (a *NetworkAdapter) processUpdateOrDeleteEvent(obj interface{}) {

	object := obj.(metav1.Object)
	if controllerRef := cloudcommon.GetControllerOf(object); controllerRef != nil {
		network := a.resolveControllerRef(object.GetNamespace(), controllerRef)
		if network == nil {
			glog.Infof("could not resolve network from obj: %s", object.GetName())
			return
		}
		key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(network)
		if err != nil {
			glog.Errorf("Network deletion state error %v", err)
			return
		}
		glog.V(4).Infof("Network %s received update event for %s %s\n", key, reflect.TypeOf(object).String(), object.GetName())
		a.queue.Add(key)
		return
	}

}
