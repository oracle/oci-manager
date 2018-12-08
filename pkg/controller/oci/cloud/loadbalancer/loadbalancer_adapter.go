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

package loadbalancer

import (
	"errors"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	cloudv1alpha1 "github.com/oracle/oci-manager/pkg/apis/cloud.k8s.io/v1alpha1"
	common "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	ocilbv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocilb.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	cloudcommon "github.com/oracle/oci-manager/pkg/controller/oci/cloud/common"
	cloudcompute "github.com/oracle/oci-manager/pkg/controller/oci/cloud/compute"
)

type LoadBalancerAdapter struct {
	subscribtions []schema.GroupVersionResource
	clientset     versioned.Interface
	lister        cache.GenericLister
	queue         workqueue.RateLimitingInterface
}

var controllerKind = cloudv1alpha1.SchemeGroupVersion.WithKind(cloudv1alpha1.LoadBalancerKind)

// init to register loadBalancer cloud type
func init() {
	cloudcommon.RegisterCloudType(
		cloudv1alpha1.LoadBalancerResourcePlural,
		cloudv1alpha1.LoadBalancerKind,
		cloudv1alpha1.GroupName,
		&cloudv1alpha1.LoadBalancerValidation,
		NewLoadBalancerAdapter,
	)
}

// factory method
func NewLoadBalancerAdapter(clientSet versioned.Interface, kubeclient kubernetes.Interface) cloudcommon.CloudTypeAdapter {
	na := LoadBalancerAdapter{
		clientset: clientSet,
	}
	na.subscribtions = subscribe()

	rand.Seed(time.Now().UTC().UnixNano())
	return &na
}

// subscribe to oci lb events
func subscribe() []schema.GroupVersionResource {
	subs := make([]schema.GroupVersionResource, 0)
	subs = append(subs, ocilbv1alpha1.SchemeGroupVersion.WithResource(ocilbv1alpha1.BackendResourcePlural))
	subs = append(subs, ocilbv1alpha1.SchemeGroupVersion.WithResource(ocilbv1alpha1.BackendSetResourcePlural))
	subs = append(subs, ocilbv1alpha1.SchemeGroupVersion.WithResource(ocilbv1alpha1.CertificateResourcePlural))
	subs = append(subs, ocilbv1alpha1.SchemeGroupVersion.WithResource(ocilbv1alpha1.ListenerResourcePlural))
	subs = append(subs, ocilbv1alpha1.SchemeGroupVersion.WithResource(ocilbv1alpha1.LoadBalancerResourcePlural))
	return subs
}

// set lister
func (a *LoadBalancerAdapter) SetLister(lister cache.GenericLister) {
	a.lister = lister
}

// set queue
func (a *LoadBalancerAdapter) SetQueue(q workqueue.RateLimitingInterface) {
	a.queue = q
}

// kind
func (a *LoadBalancerAdapter) Kind() string {
	return cloudv1alpha1.LoadBalancerKind
}

// resource
func (a *LoadBalancerAdapter) Resource() string {
	return cloudv1alpha1.LoadBalancerResourcePlural
}

// group version with resource
func (a *LoadBalancerAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return cloudv1alpha1.SchemeGroupVersion.WithResource(cloudv1alpha1.LoadBalancerResourcePlural)
}

// subscriptions
func (a *LoadBalancerAdapter) Subscriptions() []schema.GroupVersionResource {
	return a.subscribtions
}

// object meta
func (a *LoadBalancerAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*cloudv1alpha1.LoadBalancer).ObjectMeta
}

// equivalent
func (a *LoadBalancerAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	lb1 := obj1.(*cloudv1alpha1.LoadBalancer)
	if obj2 == nil {
		return false
	}
	lb2 := obj2.(*cloudv1alpha1.LoadBalancer)

	status_equal := reflect.DeepEqual(lb1.Status, lb2.Status)
	glog.Infof("Equivalent Status: %v", status_equal)

	spec_equal := reflect.DeepEqual(lb1.Spec, lb2.Spec)
	glog.Infof("Equivalent Spec: %v", spec_equal)

	equal := spec_equal && status_equal &&
		lb1.Status.State != cloudv1alpha1.OperatorStatePending

	glog.Infof("Equivalent: %v", equal)

	return equal
}

// delete loadBalancer cloud event handler
func (a *LoadBalancerAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	lb := obj.(*cloudv1alpha1.LoadBalancer)

	wait := false

	// Delete the listeners
	for _, listener := range lb.Spec.Listeners {
		listenerName := lb.Name + strconv.Itoa(listener.Port)
		glog.Infof("deleting listener: %s", listenerName)

		listenerResp, err := DeleteListener(a.clientset, lb.Namespace, listenerName)
		if err != nil {
			cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.ListenerKind, err.Error())
			return lb, err
		}
		if listenerResp != nil {
			err = errors.New("listener resource is not deleted yet")
			cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.ListenerKind, err.Error())
			wait = true
		}
	}

	// delete backends based on a selector
	listOptions := metav1.ListOptions{LabelSelector: cloudv1alpha1.LoadBalancerKind + "=" + lb.Name}
	backends, err := a.clientset.OcilbV1alpha1().Backends(lb.Namespace).List(listOptions)
	glog.Infof("DeleteLoadBalancer: backends: %v", backends.Items)
	for _, backend := range backends.Items {
		glog.Infof("deleting backend: %s", backend.Name)
		// Delete the backends
		bs, err := DeleteBackend(a.clientset, lb.Namespace, backend.Name)
		if err != nil {
			cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.BackendKind, err.Error())
			return lb, err
		}
		if bs != nil {
			err = errors.New("backend resource is not deleted yet")
			cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.BackendKind, err.Error())
			wait = true
		}
	}

	// Delete the backendset resource
	bs, err := DeleteBackendSet(a.clientset, lb)
	if err != nil {
		cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.BackendSetKind, err.Error())
		return lb, err
	}
	if bs != nil {
		err = errors.New("backendset resource is not deleted yet")
		cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.BackendSetKind, err.Error())
		wait = true
	}

	// Delete the loadbalancer resource
	lbr, err := DeleteLoadBalancer(a.clientset, lb)
	if err != nil {
		cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.LoadBalancerKind, err.Error())
		return lb, err
	}
	if lbr != nil {
		err = errors.New("loadbalancer resource is not deleted yet")
		cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.LoadBalancerKind, err.Error())
		wait = true
	}

	subnetOffset, err := cloudcommon.GetSubnetOffset(a.clientset, lb.Namespace, lb.Status.Network, lb.Name)
	if err != nil {
		glog.Errorf("error getting subnet offset: %v", err)
		cloudcommon.SetCondition(&lb.Status.OperatorStatus, "GetSubnetOffset", err.Error())
		return lb, err
	}

	// Delete the subnet resource
	for _, availabilityDomain := range lb.Status.AvailabilityZones {
		glog.Infof("deleting subnet for ad: %s", availabilityDomain)
		subnetName := lb.Name + "-" + strconv.Itoa(subnetOffset+1)
		subnet, err := cloudcompute.DeleteSubnet(a.clientset, lb.Namespace, subnetName)
		if err != nil {
			cloudcommon.SetCondition(&lb.Status.OperatorStatus, v1alpha1.SubnetKind, err.Error())
			return lb, err
		}
		if subnet != nil {
			err = errors.New("subnet resources are not deleted yet")
			cloudcommon.SetCondition(&lb.Status.OperatorStatus, v1alpha1.SubnetKind, err.Error())
			wait = true
		}
	}

	// Remove finalizers
	if !wait && len(lb.Finalizers) > 0 {
		lb.SetFinalizers([]string{})
		return lb, nil
	}

	return lb, nil
}

// update loadBalancer cloud event handler
func (a *LoadBalancerAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	loadBalancer := obj.(*cloudv1alpha1.LoadBalancer)
	resultObj, e := a.clientset.CloudV1alpha1().LoadBalancers(loadBalancer.Namespace).Update(loadBalancer)
	return resultObj, e
}

// get array of instance names for a given compute
func getInstanceNames(compute cloudv1alpha1.Compute) []string {
	instances := make([]string, compute.Spec.Replicas)
	for ord := 0; ord < compute.Spec.Replicas; ord++ {
		instances = append(instances, fmt.Sprintf("%s-%d", compute.Name, ord))
	}
	return instances
}

// reconcile - handles create and updates
func (a *LoadBalancerAdapter) Reconcile(obj runtime.Object) (runtime.Object, error) {
	lb := obj.(*cloudv1alpha1.LoadBalancer)

	// default to a http 80 "GET /" health check
	if lb.Spec.HealthCheck.Protocol == "" {
		lb.Spec.HealthCheck.Protocol = "HTTP"
	}
	if lb.Spec.HealthCheck.Port == 0 {
		lb.Spec.HealthCheck.Port = 80
	}
	if lb.Spec.HealthCheck.URLPath == "" {
		lb.Spec.HealthCheck.URLPath = "/"
	}

	if lb.Spec.BalanceMode == "" {
		lb.Spec.BalanceMode = "ROUND_ROBIN"
	}

	if lb.Spec.BandwidthMbps == "" {
		lb.Spec.BandwidthMbps = "100Mbps"
	}

	for _, listener := range lb.Spec.Listeners {
		if listener.Protocol == "" {
			listener.Protocol = "HTTP"
		}
		if listener.Port == 0 {
			listener.Port = 80
		}
		if listener.IdleTimeoutSec == 0 {
			listener.IdleTimeoutSec = 300
		}

	}

	controllerRef := cloudcommon.CreateControllerRef(lb, controllerKind)

	reconcileState := cloudv1alpha1.OperatorStateCreated

	// 1-time get from compartment and randomize
	var availabilityDomains []string
	if len(lb.Status.AvailabilityZones) == 0 {

		ads, err := cloudcommon.GetAvailabilityDomains(a.clientset, lb.Namespace, lb.Namespace)
		if err != nil {
			cloudcommon.SetCondition(&lb.Status.OperatorStatus, "AvailabilityDomain", err.Error())
			return lb, err
		}
		availabilityDomains = make([]string, len(ads))
		perm := rand.Perm(len(ads))
		for i, v := range perm {
			availabilityDomains[v] = ads[i]
		}
		lb.Status.AvailabilityZones = availabilityDomains
		glog.Infof("ad set: %v", availabilityDomains)

	} else {
		availabilityDomains = lb.Status.AvailabilityZones
	}

	lb.Status.Instances = 0
	lb.Status.AvailableInstances = 0
	selector := ""
	for k, v := range lb.Spec.ComputeSelector {
		if selector != "" {
			selector += ","
		}
		selector += k + "=" + v
	}
	listOptions := metav1.ListOptions{LabelSelector: selector}
	computes, err := a.clientset.CloudV1alpha1().Computes(lb.Namespace).List(listOptions)
	if err != nil {
		return nil, err
	}
	glog.Infof("ComputeSelector: %s matched: %v computes", selector, len(computes.Items))

	var network string
	if lb.Status.Network != "" {
		network = lb.Status.Network
	} else {
		for _, compute := range computes.Items {
			if network != "" && compute.Spec.Network != network {
				errMsg := "mismatch compute: " + compute.Name + " has network " + compute.Spec.Network +
					" - previous matching compute has: " + network
				return nil, errors.New(errMsg)
			}
			network = compute.Spec.Network
		}
		lb.Status.Network = network
	}

	instances := make(map[string]int)
	for _, compute := range computes.Items {
		weight := 1
		lb.Status.Instances += compute.Spec.Replicas

		for k, v := range lb.Spec.LabelWeightMap {
			for _, labelValue := range compute.Labels {
				if k == labelValue {
					weight = v
					break
				}
			}
		}
		if len(compute.Status.AvailabilityZones) == 0 {
			return nil, errors.New("no availabilityzones on compute yet")
		}
		for _, i := range getInstanceNames(compute) {
			if i == "" {
				continue
			}
			instances[i] = weight
		}
	}
	glog.Infof("lb instances: %v", instances)

	subnetOffset, err := cloudcommon.GetSubnetOffset(a.clientset, lb.Namespace, network, lb.Name)
	if err != nil {
		return nil, err
	}

	allSubnetsReady := true
	subnets := make([]string, 0)
	networkObj, err := a.clientset.CloudV1alpha1().Networks(lb.Namespace).Get(network, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("error getting network: %v", err)
	}
	// happens when default hasn't been populated back into spec
	if networkObj.Spec.CidrBlock == "" {
		cloudcommon.SetCondition(&lb.Status.OperatorStatus, "GetNetwork", "cidr on network is empty")
		return lb, err
	} else {
		cloudcommon.RemoveCondition(&lb.Status.OperatorStatus, "GetNetwork")
	}
	networkOctets := strings.Split(networkObj.Spec.CidrBlock, ".")

	// Process the subnet and instance resources
	for i, availabilityDomain := range availabilityDomains {
		subnetOctet := subnetOffset + i + 1

		cidrBlock := networkOctets[0] + "." + networkOctets[1] + "." + strconv.Itoa(subnetOctet) + ".0/24"
		subnetName := strconv.Itoa(subnetOctet)
		subnet, _, err := cloudcompute.CreateOrUpdateSubnet(
			a.clientset,
			lb.Namespace,
			cloudv1alpha1.LoadBalancerKind,
			lb.Name,
			network,
			controllerRef,
			subnetName,
			availabilityDomain,
			cidrBlock,
			&lb.Spec.SecuritySelector)

		subnets = append(subnets, lb.Name+"-"+subnetName)

		if err != nil {
			cloudcommon.SetCondition(&lb.Status.OperatorStatus, v1alpha1.SubnetKind, err.Error())
			return lb, nil
		}

		if subnet.Status.State != common.ResourceStateProcessed || !subnet.IsResource() {
			err = errors.New("subnet resources are not processed yet")
			cloudcommon.SetCondition(&lb.Status.OperatorStatus, v1alpha1.SubnetKind, err.Error())
			reconcileState = cloudv1alpha1.OperatorStatePending
			allSubnetsReady = false
		}
	}
	if !allSubnetsReady {
		glog.Infof("returning due to all subnets not ready")
		return lb, nil
	}

	// process the loadbalancer resource
	lbResource, _, err := CreateOrUpdateLoadBalancer(a.clientset, controllerRef, lb, subnets)
	if err != nil {
		cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.LoadBalancerKind, err.Error())
		return lb, nil
	}
	if lbResource.Status.State != common.ResourceStateProcessed || !lbResource.IsResource() {
		err = errors.New("LoadBalancer resource is not processed yet")
		cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.LoadBalancerKind, err.Error())
		lb.Status.State = cloudv1alpha1.OperatorStatePending
	} else {
		cloudcommon.RemoveCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.LoadBalancerKind)
		if lbResource.Status.Resource != nil {
			lb.Status.IPAddress = *lbResource.Status.Resource.IpAddresses[0].IpAddress
		}
	}

	// process the backendset resource
	bs, _, err := CreateOrUpdateBackendSet(a.clientset, controllerRef, lb)
	if err != nil {
		cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.BackendSetKind, err.Error())
		return lb, nil
	}
	if bs.Status.State != common.ResourceStateProcessed || !bs.IsResource() {
		err = errors.New("BackendSet resource is not processed yet")
		cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.BackendSetKind, err.Error())
		lb.Status.State = cloudv1alpha1.OperatorStatePending
	} else {
		cloudcommon.RemoveCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.BackendSetKind)
	}

	// process instances/backends
	for instance, weight := range instances {
		be, _, err := CreateOrUpdateBackend(a.clientset, controllerRef, lb, instance, weight)
		if err != nil {
			cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.BackendKind, err.Error())
			return lb, nil
		}
		if be.Status.State != common.ResourceStateProcessed || !be.IsResource() {
			err = errors.New("Backend resource is not processed yet")
			cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.BackendKind, err.Error())
			lb.Status.State = cloudv1alpha1.OperatorStatePending
		} else {
			cloudcommon.RemoveCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.BackendKind)
		}
	}

	// remove orphan instances that do not match anymore
	selector = cloudv1alpha1.LoadBalancerKind + "=" + lb.Name
	listOptions = metav1.ListOptions{LabelSelector: selector}
	backends, err := a.clientset.OcilbV1alpha1().Backends(lb.Namespace).List(listOptions)
	if err != nil {
		glog.Errorf("error listing backends with selector: %s - err: %v", selector, err)
		return nil, nil
	}
	glog.Infof("current selector: %s, backends: %v", selector, backends.Items)
	for _, be := range backends.Items {
		_, present := instances[be.Spec.InstanceRef]
		if !present {
			glog.Infof("deleting orphan backend: %s", be.Name)
			err = a.clientset.OcilbV1alpha1().Backends(lb.Namespace).Delete(be.Name, nil)
			if err != nil {
				glog.Errorf("error deleting backend: %s - %v", be.Name, err)
			}
		}
	}

	// process the listener resources
	for _, l := range lb.Spec.Listeners {

		if l.SSLCertificate.Certificate != "" {
			cert, _, err := CreateOrUpdateCertificate(a.clientset, controllerRef, lb, &l)
			if err != nil {
				cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.CertificateKind, err.Error())
				return lb, nil
			}
			if cert.Status.State != common.ResourceStateProcessed || !cert.IsResource() {
				err = errors.New("Certificate resource is not processed yet")
				cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.CertificateKind, err.Error())
				lb.Status.State = cloudv1alpha1.OperatorStatePending
				return lb, nil
			} else {
				cloudcommon.RemoveCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.CertificateKind)
			}
		}

		listener, _, err := CreateOrUpdateListener(a.clientset, controllerRef, lb, &l)
		if err != nil {
			cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.ListenerKind, err.Error())
			return lb, nil
		}
		if listener.Status.State != common.ResourceStateProcessed || !listener.IsResource() {
			err = errors.New("Listener resource is not processed yet")
			cloudcommon.SetCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.ListenerKind, err.Error())
			lb.Status.State = cloudv1alpha1.OperatorStatePending
			return lb, nil
		} else {
			cloudcommon.RemoveCondition(&lb.Status.OperatorStatus, ocilbv1alpha1.ListenerKind)
		}
	}

	// Everything is done. Update the State, reset the Conditions and return
	if reconcileState == cloudv1alpha1.OperatorStateCreated {
		lb.Status.State = cloudv1alpha1.OperatorStateCreated
		if lb.Status.Conditions != nil {
			lb.Status.Conditions = []cloudv1alpha1.OperatorCondition{}
		}
	} else {
		lb.Status.State = cloudv1alpha1.OperatorStatePending
	}
	return lb, nil

}

// callback for resource
func (a *LoadBalancerAdapter) CallbackForResource(resource schema.GroupVersionResource) cache.ResourceEventHandlerFuncs {
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
func (a *LoadBalancerAdapter) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *cloudv1alpha1.LoadBalancer {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}
	obj, err := a.lister.ByNamespace(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}

	loadBalancer := obj.(*cloudv1alpha1.LoadBalancer)

	if loadBalancer.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return loadBalancer
}

// ignore add event
func (a *LoadBalancerAdapter) ignoreAddEvent(obj interface{}) {
	glog.V(4).Infof("Got add event: %v", obj)
}

// handle update or delete events
func (a *LoadBalancerAdapter) processUpdateOrDeleteEvent(obj interface{}) {
	object := obj.(metav1.Object)
	if controllerRef := cloudcommon.GetControllerOf(object); controllerRef != nil {
		loadBalancer := a.resolveControllerRef(object.GetNamespace(), controllerRef)
		if loadBalancer == nil {
			return
		}
		key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(loadBalancer)
		if err != nil {
			glog.Errorf("LoadBalancer deletion state error %v", err)
			return
		}
		glog.V(4).Infof("LoadBalancer %s received update event for %s %s\n", key, reflect.TypeOf(object).String(), object.GetName())
		a.queue.Add(key)
		//a.Reconcile(loadBalancer)
		return
	}
}
