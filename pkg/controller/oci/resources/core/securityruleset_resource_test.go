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
	"testing"

	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	fakeoci "github.com/oracle/oci-manager/pkg/controller/oci/resources/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	fakeclient "github.com/oracle/oci-manager/pkg/client/clientset/versioned/fake"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
)

var (
	securityRuleSettest1 = corev1alpha1.SecurityRuleSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "securityRuleSet.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       corev1alpha1.SecurityRuleSetKind,
		},
		Spec: corev1alpha1.SecurityRuleSetSpec{
			DisplayName:    "bla",
			VcnRef:         "vcn.test1",
			CompartmentRef: "compartment.test1",
		},
	}
)

func TestSecurityRuleSetResourceBasic(t *testing.T) {

	t.Log("Testing securityRuleSet_resource")
	clientset := fakeclient.NewSimpleClientset()
	vcnClient := fakeoci.NewVcnClient()

	securityRuleSetAdapter := SecurityRuleSetAdapter{}
	securityRuleSetAdapter.clientset = clientset
	securityRuleSetAdapter.vcnClient = vcnClient

	comp, err := clientset.OciidentityV1alpha1().Compartments(fakeNs).Create(&compartment)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created compartment object %v", comp)

	vcn, err := clientset.OcicoreV1alpha1().Vcns(fakeNs).Create(&vcntest1)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created vcn object %v", vcn)

	newSecurityRuleSet, err := securityRuleSetAdapter.CreateObject(&securityRuleSettest1)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created SecurityRuleSet object %v", newSecurityRuleSet)

	if securityRuleSetAdapter.IsExpectedType(newSecurityRuleSet) {
		t.Logf("Checked SecurityRuleSet type")
	}

	securityRuleSetWithResource, err := securityRuleSetAdapter.Create(newSecurityRuleSet)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created SecurityRuleSet resource %v", securityRuleSetWithResource)

	securityRuleSetWithResource, err = securityRuleSetAdapter.Get(newSecurityRuleSet)
	securityRuleSetWithResource, err = securityRuleSetAdapter.Update(newSecurityRuleSet)
	securityRuleSetWithResource, err = securityRuleSetAdapter.Delete(newSecurityRuleSet)
	eq := securityRuleSetAdapter.Equivalent(newSecurityRuleSet, newSecurityRuleSet)
	if !eq {
		t.Errorf("should be equal")
	}
	_, err = securityRuleSetAdapter.DependsOnRefs(newSecurityRuleSet)
}

func NewFakeSecurityRuleSetAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	vcnClient := fakeoci.NewVcnClient()
	securityRuleSetAdapter := SecurityRuleSetAdapter{}
	securityRuleSetAdapter.clientset = clientset
	securityRuleSetAdapter.vcnClient = vcnClient
	return &securityRuleSetAdapter
}
