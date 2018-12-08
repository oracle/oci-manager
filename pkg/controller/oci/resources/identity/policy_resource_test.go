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

package identity

import (
	"testing"

	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	fakeoci "github.com/oracle/oci-manager/pkg/controller/oci/resources/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ocisdkidentity "github.com/oracle/oci-go-sdk/identity"

	identityv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com/v1alpha1"
	fakeclient "github.com/oracle/oci-manager/pkg/client/clientset/versioned/fake"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
)

var (
	compartment = identityv1alpha1.Compartment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "compartment.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       identityv1alpha1.CompartmentKind,
		},
		Spec: identityv1alpha1.CompartmentSpec{
			Description: "testCompartment"},
		Status: identityv1alpha1.CompartmentStatus{
			Resource: &identityv1alpha1.CompartmentResource{
				Compartment: ocisdkidentity.Compartment{Id: resourcescommon.StrPtrOrNil("fakeCompartmentOCIID")},
			},
		},
	}

	desc = "bla"

	policyResource = identityv1alpha1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "policy.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ociidentity.oracle.com/v1alpha1",
			Kind:       identityv1alpha1.PolicyKind,
		},
		Spec: identityv1alpha1.PolicySpec{
			CompartmentRef: "compartment.test1",
			Description:    &desc,
		},
	}
)

func TestPolicyResourceBasic(t *testing.T) {

	t.Log("Testing policy_resource")
	clientset := fakeclient.NewSimpleClientset()
	idClient := fakeoci.NewIdentityClient()

	policyAdapter := PolicyAdapter{}
	policyAdapter.clientset = clientset
	policyAdapter.idClient = idClient

	comp, err := clientset.OciidentityV1alpha1().Compartments(fakeNs).Create(&compartment)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created compartment object %v", comp)

	newPolicy, err := policyAdapter.CreateObject(&policyResource)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Policy object %v", newPolicy)

	if policyAdapter.IsExpectedType(newPolicy) {
		t.Logf("Checked Policy type")
	}

	policyWithResource, err := policyAdapter.Create(newPolicy)

	if err != nil {
		t.Errorf("Got create policy error %v", err)
	}
	t.Logf("Created Policy resource %v", policyWithResource)

	policyWithResource, err = policyAdapter.Get(newPolicy)
	policyWithResource, err = policyAdapter.Update(newPolicy)
	policyWithResource, err = policyAdapter.Delete(newPolicy)
	eq := policyAdapter.Equivalent(newPolicy, newPolicy)
	if !eq {
		t.Errorf("should be equal")
	}
	_, err = policyAdapter.DependsOnRefs(newPolicy)
}

func NewFakePolicyAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	idClient := fakeoci.NewIdentityClient()
	policyAdapter := PolicyAdapter{}
	policyAdapter.clientset = clientset
	policyAdapter.idClient = idClient
	return &policyAdapter
}
