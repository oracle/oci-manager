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
package security

import (
	"errors"
	"k8s.io/client-go/kubernetes"
	"reflect"

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
		cloudv1alpha1.SecurityResourcePlural,
		cloudv1alpha1.SecurityKind,
		cloudv1alpha1.GroupName,
		&cloudv1alpha1.SecurityValidation,
		NewSecurityAdapter,
	)
}

var controllerKind = cloudv1alpha1.SchemeGroupVersion.WithKind("Security")

type SecurityAdapter struct {
	subscribtions []schema.GroupVersionResource
	clientset     versioned.Interface
	lister        cache.GenericLister
	queue         workqueue.RateLimitingInterface
}

func subscribe() []schema.GroupVersionResource {
	subs := make([]schema.GroupVersionResource, 0)
	subs = append(subs, ocicorev1alpha1.SchemeGroupVersion.WithResource(ocicorev1alpha1.SecurityRuleSetResourcePlural))

	return subs
}

func NewSecurityAdapter(clientSet versioned.Interface, kubeclient kubernetes.Interface) cloudcommon.CloudTypeAdapter {
	na := SecurityAdapter{
		clientset: clientSet,
	}

	na.subscribtions = subscribe()

	return &na
}

func (a *SecurityAdapter) SetLister(lister cache.GenericLister) {
	a.lister = lister
}

func (a *SecurityAdapter) SetQueue(q workqueue.RateLimitingInterface) {
	a.queue = q
}

func (a *SecurityAdapter) Kind() string {
	return cloudv1alpha1.SecurityKind
}

func (a *SecurityAdapter) Resource() string {
	return cloudv1alpha1.SecurityResourcePlural
}

func (a *SecurityAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return cloudv1alpha1.SchemeGroupVersion.WithResource(cloudv1alpha1.SecurityResourcePlural)
}

func (a *SecurityAdapter) Subscriptions() []schema.GroupVersionResource {
	return a.subscribtions
}

func (a *SecurityAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*cloudv1alpha1.Security).ObjectMeta
}

func (a *SecurityAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	sec1 := obj1.(*cloudv1alpha1.Security)
	sec2 := obj1.(*cloudv1alpha1.Security)

	status_equal := reflect.DeepEqual(sec1.Status, sec2.Status)
	glog.Infof("Equivalent Status: %v", status_equal)

	spec_equal := reflect.DeepEqual(sec1.Spec, sec2.Spec)
	glog.Infof("Equivalent Spec: %v", spec_equal)

	equal := spec_equal && status_equal &&
		sec1.Status.State != cloudv1alpha1.OperatorStatePending

	glog.Infof("Equivalent: %v", equal)

	return equal
}

func (a *SecurityAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	security := obj.(*cloudv1alpha1.Security)

	wait := false

	selector := "security=" + security.Name
	listOptions := metav1.ListOptions{LabelSelector: selector}
	queryResult, err := a.clientset.OcicoreV1alpha1().SecurityRuleSets(security.Namespace).List(listOptions)
	if err != nil {
		return nil, err
	}
	glog.Infof("SecurityRuleSet selector: %s matched: %v", selector, queryResult.Items)

	for _, securityRuleSet := range queryResult.Items {
		// Delete the security list resources
		srs, err := DeleteSecurityRuleSet(a.clientset, security, securityRuleSet.Name)
		if err != nil {
			cloudcommon.SetCondition(&security.Status.OperatorStatus, v1alpha1.SecurityRuleSetKind, err.Error())
			return security, err
		}
		if srs != nil {
			err = errors.New("security rule set resource is not deleted yet")
			cloudcommon.SetCondition(&security.Status.OperatorStatus, v1alpha1.SecurityRuleSetKind, err.Error())
			wait = true
		}
	}

	// Remove finalizers
	if !wait && len(security.Finalizers) > 0 {
		security.SetFinalizers([]string{})
		return security, nil
	}

	return security, nil
}

func (a *SecurityAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	security := obj.(*cloudv1alpha1.Security)
	resultObj, e := a.clientset.CloudV1alpha1().Securities(security.Namespace).Update(security)
	return resultObj, e
}

func (a *SecurityAdapter) Reconcile(obj runtime.Object) (runtime.Object, error) {

	security := obj.(*cloudv1alpha1.Security)

	controllerRef := cloudcommon.CreateControllerRef(security, controllerKind)
	reconcileState := cloudv1alpha1.OperatorStateCreated

	selector := ""
	for k, v := range security.Spec.NetworkSelector {
		if selector != "" {
			selector += ","
		}
		selector += k + "=" + v
	}
	listOptions := metav1.ListOptions{LabelSelector: selector}
	nets, err := a.clientset.CloudV1alpha1().Networks(security.Namespace).List(listOptions)
	if err != nil {
		return nil, err
	}
	glog.Infof("NetworkSelector: %s matched: %v", selector, nets.Items)

	for _, net := range nets.Items {
		// Process the security list resource
		securityruleset, _, err := CreateOrUpdateSecurityRuleSet(a.clientset, security, controllerRef, net.Name)
		if err != nil {
			cloudcommon.SetCondition(&security.Status.OperatorStatus, v1alpha1.SecurityRuleSetKind, err.Error())
			return security, err
		}
		if securityruleset.Status.State != common.ResourceStateProcessed || !securityruleset.IsResource() {
			err = errors.New("security list resource is not processed yet")
			cloudcommon.SetCondition(&security.Status.OperatorStatus, v1alpha1.SecurityRuleSetKind, err.Error())
			reconcileState = cloudv1alpha1.OperatorStatePending
		} else {
			cloudcommon.RemoveCondition(&security.Status.OperatorStatus, v1alpha1.SecurityRuleSetKind)
		}
	}

	// Everything is done. Update the State, reset the Conditions and return
	if reconcileState == cloudv1alpha1.OperatorStateCreated {
		security.Status.State = cloudv1alpha1.OperatorStateCreated
		security.Status.Conditions = []cloudv1alpha1.OperatorCondition{}
	} else {
		security.Status.State = cloudv1alpha1.OperatorStatePending
	}
	return security, nil

}

func (a *SecurityAdapter) CallbackForResource(resource schema.GroupVersionResource) cache.ResourceEventHandlerFuncs {
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

func (a *SecurityAdapter) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *cloudv1alpha1.Security {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}
	obj, err := a.lister.ByNamespace(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}

	security := obj.(*cloudv1alpha1.Security)

	if security.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return security
}

func (a *SecurityAdapter) ignoreAddEvent(obj interface{}) {
	glog.V(4).Infof("Got add event: %v - type: %s", obj, reflect.TypeOf(obj))
}

func (a *SecurityAdapter) processUpdateOrDeleteEvent(obj interface{}) {

	object := obj.(metav1.Object)
	if controllerRef := cloudcommon.GetControllerOf(object); controllerRef != nil {
		security := a.resolveControllerRef(object.GetNamespace(), controllerRef)
		if security == nil {
			glog.Infof("could not resolve security from obj: %s", object.GetName())
			return
		}
		key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(security)
		if err != nil {
			glog.Errorf("Security deletion state error %v", err)
			return
		}
		glog.V(4).Infof("Security %s received update event for %s %s\n", key, reflect.TypeOf(object).String(), object.GetName())
		a.queue.Add(key)
		return
	}

}
