/*
Copyright 2018 Oracle and/or its affiliates. All rpolicyhts reserved.

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

package identity

import (
	"errors"
	"k8s.io/client-go/kubernetes"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ocicommon "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	identitygroup "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com"
	ociidentityv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com/v1alpha1"

	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"

	ocisdkcommon "github.com/oracle/oci-go-sdk/common"
	ociidentity "github.com/oracle/oci-go-sdk/identity"

	"context"

	"github.com/golang/glog"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	// "golang.org/x/tools/cmd/guru/testdata/src/alias"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

func init() {
	resourcescommon.RegisterResourceTypeWithValidation(
		identitygroup.GroupName,
		ociidentityv1alpha1.PolicyKind,
		ociidentityv1alpha1.PolicyResourcePlural,
		ociidentityv1alpha1.PolicyControllerName,
		&ociidentityv1alpha1.PolicyValidation,
		NewPolicyAdapter)
}

// PolicyAdapter implements the adapter interface for policy resource
type PolicyAdapter struct {
	clientset versioned.Interface
	ctx       context.Context
	idClient  resourcescommon.IdentityClientInterface
}

// NewPolicyAdapter creates a new adapter for policy resource
func NewPolicyAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfpolicy ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	pa := PolicyAdapter{}

	idClient, err := ociidentity.NewIdentityClientWithConfigurationProvider(ociconfpolicy)

	if err != nil {
		glog.Errorf("Error creating oci VCN client: %v", err)
		os.Exit(1)
	}

	pa.idClient = &idClient
	pa.clientset = clientset
	pa.ctx = context.Background()

	return &pa
}

// Kind returns the resource kind string
func (a *PolicyAdapter) Kind() string {
	return ociidentityv1alpha1.PolicyKind
}

// Resource returns the plural name of the resource type
func (a *PolicyAdapter) Resource() string {
	return ociidentityv1alpha1.PolicyResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *PolicyAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ociidentityv1alpha1.SchemeGroupVersion.WithResource(ociidentityv1alpha1.PolicyResourcePlural)
}

// ObjectType returns the policy type for this adapter
func (a *PolicyAdapter) ObjectType() runtime.Object {
	return &ociidentityv1alpha1.Policy{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *PolicyAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ociidentityv1alpha1.Policy)
	return ok
}

// Copy returns a copy of a policy object
func (a *PolicyAdapter) Copy(obj runtime.Object) runtime.Object {
	internetgateway := obj.(*ociidentityv1alpha1.Policy)
	return internetgateway.DeepCopyObject()
}

// Equivalent checks if two policy objects are the same
func (a *PolicyAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	policy1 := obj1.(*ociidentityv1alpha1.Policy)
	policy2 := obj2.(*ociidentityv1alpha1.Policy)
	if policy1.Status.Resource != nil {
		policy1.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	if policy2.Status.Resource != nil {
		policy2.Status.Resource.TimeCreated = &ocisdkcommon.SDKTime{}
	}
	return reflect.DeepEqual(policy1, policy2)
}

// IsResourceCompliant checks if resource config is complient with CRD spec
func (a *PolicyAdapter) IsResourceCompliant(obj runtime.Object) bool {
	policy := obj.(*ociidentityv1alpha1.Policy)

	if policy.Status.Resource == nil {
		return false
	}

	resource := policy.Status.Resource

	if resource.LifecycleState == ociidentity.PolicyLifecycleStateCreating ||
		resource.LifecycleState == ociidentity.PolicyLifecycleStateDeleting {
		return true
	}

	if resource.LifecycleState == ociidentity.PolicyLifecycleStateDeleted ||
		resource.LifecycleState == ociidentity.PolicyLifecycleStateInactive {
		return false
	}

	if *resource.Name != policy.Name ||
		resource.Description != policy.Spec.Description {
		return false
	}

	specStatements := make(map[string]bool)
	resourceStatements := make(map[string]bool)

	for _, statement := range policy.Spec.Statements {
		specStatements[statement] = true
	}
	for _, statement := range policy.Status.Resource.Statements {
		resourceStatements[statement] = true
	}

	return reflect.DeepEqual(specStatements, resourceStatements)

}

// IsResourceStatusChanged checks if two vcn objects are the same
func (a *PolicyAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	policy1 := obj1.(*ociidentityv1alpha1.Policy)
	policy2 := obj2.(*ociidentityv1alpha1.Policy)

	return policy1.Status.Resource.LifecycleState != policy2.Status.Resource.LifecycleState
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *PolicyAdapter) Id(obj runtime.Object) string {
	return obj.(*ociidentityv1alpha1.Policy).GetResourceID()
}

// ObjectMeta returns the object meta struct from the policy object
func (a *PolicyAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ociidentityv1alpha1.Policy).ObjectMeta
}

// DependsOn returns a map of policy dependencies (objects that the policy depends on)
func (a *PolicyAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ociidentityv1alpha1.Policy).Spec.DependsOn
}

// Dependents returns a map of policy dependents (objects that depend on the policy)
func (a *PolicyAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ociidentityv1alpha1.Policy).Status.Dependents
}

// CreateObject creates the policy object
func (a *PolicyAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ociidentityv1alpha1.Policy)
	return a.clientset.OciidentityV1alpha1().Policies(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the policy object
func (a *PolicyAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ociidentityv1alpha1.Policy)
	return a.clientset.OciidentityV1alpha1().Policies(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the policy object
func (a *PolicyAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var object = obj.(*ociidentityv1alpha1.Policy)
	return a.clientset.OciidentityV1alpha1().Policies(object.ObjectMeta.Namespace).Delete(object.Name, options)
}

// DependsOnRefs returns the objects that the policy depends on
func (a *PolicyAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	var policy = obj.(*ociidentityv1alpha1.Policy)
	deps := make([]runtime.Object, 0)

	if !resourcescommon.IsOcid(policy.Spec.CompartmentRef) {
		compartment, err := resourcescommon.Compartment(a.clientset, policy.ObjectMeta.Namespace, policy.Spec.CompartmentRef)
		if err != nil {
			return nil, err
		}
		deps = append(deps, compartment)
	}

	return deps, nil
}

// Create creates the policy resource in oci
func (a *PolicyAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	var (
		policy        = obj.(*ociidentityv1alpha1.Policy)
		compartmentId string
		err           error
	)

	if resourcescommon.IsOcid(policy.Spec.CompartmentRef) {
		compartmentId = policy.Spec.CompartmentRef
	} else {
		compartmentId, err = resourcescommon.CompartmentId(a.clientset, policy.ObjectMeta.Namespace, policy.Spec.CompartmentRef)
		if err != nil {
			return policy, policy.Status.HandleError(err)
		}
	}

	request := ociidentity.CreatePolicyRequest{}
	request.CompartmentId = ocisdkcommon.String(compartmentId)
	request.Description = policy.Spec.Description
	request.Name = &policy.Name
	request.Statements = policy.Spec.Statements

	request.OpcRetryToken = ocisdkcommon.String(string(policy.UID))

	r, err := a.idClient.CreatePolicy(a.ctx, request)

	if err != nil {
		return policy, policy.Status.HandleError(err)
	}

	return policy.SetResource(&r.Policy), policy.Status.HandleError(err)
}

// Delete deletes the policy resource in oci
func (a *PolicyAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ociidentityv1alpha1.Policy)

	request := ociidentity.DeletePolicyRequest{
		PolicyId: object.Status.Resource.Id,
	}

	_, e := a.idClient.DeletePolicy(a.ctx, request)

	if e == nil && object.Status.Resource != nil {
		object.Status.Resource.Id = ocisdkcommon.String("")
	}
	return object, object.Status.HandleError(e)
}

// Get retrieves the policy resource from oci
func (a *PolicyAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ociidentityv1alpha1.Policy)

	request := ociidentity.GetPolicyRequest{
		PolicyId: object.Status.Resource.Id,
	}

	e := wait.PollImmediate(1*time.Second, 10*time.Second, func() (bool, error) {
		r, e := a.idClient.GetPolicy(a.ctx, request)
		if e != nil {
			return false, e
		}
		if r.LifecycleState != ociidentity.PolicyLifecycleStateCreating {
			object.SetResource(&r.Policy)
			return true, nil
		}
		return false, e
	})

	return object, object.Status.HandleError(e)
}

// Update updates the policy resource in oci
func (a *PolicyAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ociidentityv1alpha1.Policy)

	details := ociidentity.UpdatePolicyDetails{
		Description: object.Spec.Description,
		Statements:  object.Spec.Statements,
	}
	request := ociidentity.UpdatePolicyRequest{
		PolicyId:            object.Status.Resource.Id,
		UpdatePolicyDetails: details,
	}

	if object.Status.Resource.LifecycleState != ociidentity.PolicyLifecycleStateActive {
		return object, errors.New(string(object.Status.Resource.LifecycleState))
	}

	r, e := a.idClient.UpdatePolicy(a.ctx, request)

	if e != nil {
		return object, object.Status.HandleError(e)
	}

	return object.SetResource(&r.Policy), object.Status.HandleError(e)
}

// UpdateForResource calls a common UpdateForResource method to update the policy resource in the policy object
func (a *PolicyAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
