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

package core

import (
	"errors"
	"k8s.io/client-go/kubernetes"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ocicommon "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	coregroup "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com"
	ocicorev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"

	ocisdkcommon "github.com/oracle/oci-go-sdk/common"
	ocicore "github.com/oracle/oci-go-sdk/core"

	"context"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

func init() {
	resourcescommon.RegisterResourceTypeWithValidation(
		coregroup.GroupName,
		ocicorev1alpha1.SecurityRuleSetKind,
		ocicorev1alpha1.SecurityRuleSetResourcePlural,
		ocicorev1alpha1.SecurityRuleSetControllerName,
		&ocicorev1alpha1.SecurityRuleSetValidation,
		NewSecurityRuleSetAdapter)
}

// SecurityRuleSetAdapter implements the adapter interface for security rule set resource
type SecurityRuleSetAdapter struct {
	clientset versioned.Interface
	ctx       context.Context
	vcnClient resourcescommon.VcnClientInterface
}

// NewSecurityRuleSetAdapter creates a new adapter for security rule set resource
func NewSecurityRuleSetAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	sla := SecurityRuleSetAdapter{}

	vcnClient, err := ocicore.NewVirtualNetworkClientWithConfigurationProvider(ociconfig)

	if err != nil {
		glog.Errorf("Error creating oci VCN client: %v", err)
		os.Exit(1)
	}

	sla.vcnClient = &vcnClient
	sla.clientset = clientset
	sla.ctx = context.Background()
	return &sla
}

// Kind returns the resource kind string
func (a *SecurityRuleSetAdapter) Kind() string {
	return ocicorev1alpha1.SecurityRuleSetKind
}

// Resource returns the plural name of the resource type
func (a *SecurityRuleSetAdapter) Resource() string {
	return ocicorev1alpha1.SecurityRuleSetResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *SecurityRuleSetAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ocicorev1alpha1.SchemeGroupVersion.WithResource(ocicorev1alpha1.SecurityRuleSetResourcePlural)
}

// ObjectType returns the security rule set type for this adapter
func (a *SecurityRuleSetAdapter) ObjectType() runtime.Object {
	return &ocicorev1alpha1.SecurityRuleSet{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *SecurityRuleSetAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ocicorev1alpha1.SecurityRuleSet)
	return ok
}

// Copy returns a copy of a security rule set object
func (a *SecurityRuleSetAdapter) Copy(obj runtime.Object) runtime.Object {
	securityruleset := obj.(*ocicorev1alpha1.SecurityRuleSet)
	return securityruleset.DeepCopyObject()
}

// Equivalent checks if two security rule set objects are the same
func (a *SecurityRuleSetAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	securityruleset1 := obj1.(*ocicorev1alpha1.SecurityRuleSet)
	securityruleset2 := obj2.(*ocicorev1alpha1.SecurityRuleSet)
	if securityruleset1.Status.Resource != nil {
		securityruleset1.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	if securityruleset2.Status.Resource != nil {
		securityruleset2.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}

		if len(securityruleset2.Spec.EgressSecurityRules) == len(securityruleset2.Status.Resource.EgressSecurityRules) {
			for i, specRule := range securityruleset2.Spec.EgressSecurityRules {
				statusRule := securityruleset2.Status.Resource.EgressSecurityRules[i]
				statusRule.DestinationType = ""
				if specRule.IsStateless == nil {
					falseValue := false
					specRule.IsStateless = &falseValue
				}
				if !reflect.DeepEqual(specRule, statusRule) {
					glog.Infof("securityruleset equal false due to egress !reflect.DeepEqual(%v, %v)", specRule, statusRule)
					return false
				}
			}
		} else {
			glog.Infof("securityruleset equal false due to egress len")
			return false
		}

		if len(securityruleset2.Spec.IngressSecurityRules) == len(securityruleset2.Status.Resource.IngressSecurityRules) {
			for i, specRule := range securityruleset2.Spec.IngressSecurityRules {
				statusRule := securityruleset2.Status.Resource.IngressSecurityRules[i]
				statusRule.SourceType = ""
				if specRule.IsStateless == nil {
					falseValue := false
					specRule.IsStateless = &falseValue
				}
				if !reflect.DeepEqual(specRule, statusRule) {
					glog.Infof("securityruleset equal false due to ingress !reflect.DeepEqual(%v, %v)", specRule, statusRule)
					return false
				}
			}
		} else {
			glog.Infof("securityruleset equal false due to ingress len")
			return false
		}
		return true
	} else {
		return reflect.DeepEqual(securityruleset1, securityruleset2)
	}
}

// IsResourceCompliant checks if resource config is complient with CRD spec
func (a *SecurityRuleSetAdapter) IsResourceCompliant(obj runtime.Object) bool {

	securityruleset := obj.(*ocicorev1alpha1.SecurityRuleSet)

	if securityruleset.Status.Resource == nil {
		return false
	}

	resource := securityruleset.Status.Resource

	if resource.LifecycleState == ocicore.SecurityListLifecycleStateProvisioning ||
		resource.LifecycleState == ocicore.SecurityListLifecycleStateTerminating {
		return true
	}

	if resource.LifecycleState == ocicore.SecurityListLifecycleStateTerminated {
		return false
	}

	if len(securityruleset.Spec.EgressSecurityRules) == len(securityruleset.Status.Resource.EgressSecurityRules) {
		for i, specRule := range securityruleset.Spec.EgressSecurityRules {
			statusRule := securityruleset.Status.Resource.EgressSecurityRules[i]
			statusRule.DestinationType = ""
			if specRule.IsStateless == nil {
				falseValue := false
				specRule.IsStateless = &falseValue
			}
			if !reflect.DeepEqual(specRule, statusRule) {
				glog.V(5).Infof("securityruleset incomplient due to egress !reflect.DeepEqual(%v, %v)", specRule, statusRule)
				return false
			}
		}
	} else {
		glog.V(5).Infof("securityruleset incomplient due to egress len")
		return false
	}

	if len(securityruleset.Spec.IngressSecurityRules) == len(securityruleset.Status.Resource.IngressSecurityRules) {
		for i, specRule := range securityruleset.Spec.IngressSecurityRules {
			statusRule := securityruleset.Status.Resource.IngressSecurityRules[i]
			statusRule.SourceType = ""
			if specRule.IsStateless == nil {
				falseValue := false
				specRule.IsStateless = &falseValue
			}
			if !reflect.DeepEqual(specRule, statusRule) {
				glog.Infof("securityruleset incomplient due to ingress !reflect.DeepEqual(%v, %v)", specRule, statusRule)
				return false
			}
		}
	} else {
		glog.Infof("securityruleset incomplient due to ingress len")
		return false
	}
	return true

}

// IsResourceStatusChanged checks if two vcn objects are the same
func (a *SecurityRuleSetAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	securityruleset1 := obj1.(*ocicorev1alpha1.SecurityRuleSet)
	securityruleset2 := obj2.(*ocicorev1alpha1.SecurityRuleSet)

	return securityruleset1.Status.Resource.LifecycleState != securityruleset2.Status.Resource.LifecycleState
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *SecurityRuleSetAdapter) Id(obj runtime.Object) string {
	return obj.(*ocicorev1alpha1.SecurityRuleSet).GetResourceID()
}

// ObjectMeta returns the object meta struct from the security rule set object
func (a *SecurityRuleSetAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ocicorev1alpha1.SecurityRuleSet).ObjectMeta
}

// DependsOn returns a map of security rule set dependencies (objects that the security rule set depends on)
func (a *SecurityRuleSetAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ocicorev1alpha1.SecurityRuleSet).Spec.DependsOn
}

// Dependents returns a map of security rule set dependents (objects that depend on the security rule set)
func (a *SecurityRuleSetAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ocicorev1alpha1.SecurityRuleSet).Status.Dependents
}

// CreateObject creates the security rule set object
func (a *SecurityRuleSetAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.SecurityRuleSet)
	return a.clientset.OcicoreV1alpha1().SecurityRuleSets(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the security rule set object
func (a *SecurityRuleSetAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.SecurityRuleSet)
	return a.clientset.OcicoreV1alpha1().SecurityRuleSets(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the security rule set object
func (a *SecurityRuleSetAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var object = obj.(*ocicorev1alpha1.SecurityRuleSet)
	return a.clientset.OcicoreV1alpha1().SecurityRuleSets(object.ObjectMeta.Namespace).Delete(object.Name, options)
}

// DependsOnRefs returns the objects that the security rule set depends on
func (a *SecurityRuleSetAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.SecurityRuleSet)
	deps := make([]runtime.Object, 0)

	if !resourcescommon.IsOcid(object.Spec.CompartmentRef) {
		compartment, err := resourcescommon.Compartment(a.clientset, object.ObjectMeta.Namespace, object.Spec.CompartmentRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, compartment)
	}

	if !resourcescommon.IsOcid(object.Spec.VcnRef) {
		virtualnetwork, err := resourcescommon.Vcn(a.clientset, object.ObjectMeta.Namespace, object.Spec.VcnRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, virtualnetwork)
	}

	return deps, nil
}

// Create creates the security rule set resource in oci
func (a *SecurityRuleSetAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	var (
		object           = obj.(*ocicorev1alpha1.SecurityRuleSet)
		compartmentId    string
		virtualnetworkId string
		err              error
	)

	if resourcescommon.IsOcid(object.Spec.CompartmentRef) {
		compartmentId = object.Spec.CompartmentRef
	} else {
		compartmentId, err = resourcescommon.CompartmentId(a.clientset, object.ObjectMeta.Namespace, object.Spec.CompartmentRef)
		if err != nil {
			return object, object.Status.HandleError(err)
		}
	}

	if resourcescommon.IsOcid(object.Spec.VcnRef) {
		virtualnetworkId = object.Spec.VcnRef
	} else {
		virtualnetworkId, err = resourcescommon.VcnId(a.clientset, object.ObjectMeta.Namespace, object.Spec.VcnRef)
		if err != nil {
			return object, object.Status.HandleError(err)
		}
	}

	request := ocicore.CreateSecurityListRequest{}
	request.CompartmentId = ocisdkcommon.String(compartmentId)
	request.VcnId = ocisdkcommon.String(virtualnetworkId)
	request.DisplayName = resourcescommon.Display(object.Name, object.Spec.DisplayName)
	request.EgressSecurityRules = object.Spec.EgressSecurityRules
	request.IngressSecurityRules = object.Spec.IngressSecurityRules
	request.OpcRetryToken = ocisdkcommon.String(string(object.UID))

	glog.Infof("SecurityList: %s OpcRetryToken: %s", object.Name, string(object.UID))

	r, err := a.vcnClient.CreateSecurityList(a.ctx, request)

	if err != nil {
		return object, object.Status.HandleError(err)
	}

	return object.SetResource(&r.SecurityList), object.Status.HandleError(err)

}

// Delete deletes the security rule set resource in oci
func (a *SecurityRuleSetAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.SecurityRuleSet)

	request := ocicore.DeleteSecurityListRequest{
		SecurityListId: object.Status.Resource.Id,
	}

	_, e := a.vcnClient.DeleteSecurityList(a.ctx, request)

	if e == nil && object.Status.Resource != nil {
		object.Status.Resource.Id = ocisdkcommon.String("")
	}
	return object, object.Status.HandleError(e)

}

// Get retrieves the security rule set resource from oci
func (a *SecurityRuleSetAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.SecurityRuleSet)

	request := ocicore.GetSecurityListRequest{
		SecurityListId: object.Status.Resource.Id,
	}

	e := wait.PollImmediate(1*time.Second, 10*time.Second, func() (bool, error) {
		r, e := a.vcnClient.GetSecurityList(a.ctx, request)
		if e != nil {
			return false, e
		}
		if r.LifecycleState != ocicore.SecurityListLifecycleStateProvisioning {
			object.SetResource(&r.SecurityList)
			return true, nil
		}
		return false, e
	})

	return object, object.Status.HandleError(e)

}

// Update updates the security rule set resource in oci
func (a *SecurityRuleSetAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicorev1alpha1.SecurityRuleSet)

	request := ocicore.UpdateSecurityListRequest{
		SecurityListId: object.Status.Resource.Id,
	}

	request.EgressSecurityRules = object.Spec.EgressSecurityRules
	request.IngressSecurityRules = object.Spec.IngressSecurityRules

	if object.Status.Resource.LifecycleState != ocicore.SecurityListLifecycleStateAvailable {
		return object, errors.New(string(object.Status.Resource.LifecycleState))
	}

	r, e := a.vcnClient.UpdateSecurityList(a.ctx, request)

	if e != nil {
		return object, object.Status.HandleError(e)
	}

	return object.SetResource(&r.SecurityList), object.Status.HandleError(e)

}

// UpdateForResource calls a common UpdateForResource method to update the security rule set resource in the security rule set object
func (a *SecurityRuleSetAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
