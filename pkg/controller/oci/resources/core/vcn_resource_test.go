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
	corev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	identityv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com/v1alpha1"
	fakeclient "github.com/oracle/oci-manager/pkg/client/clientset/versioned/fake"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"

	ocisdkcore "github.com/oracle/oci-go-sdk/core"
	ocisdkidentity "github.com/oracle/oci-go-sdk/identity"

	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	fakeoci "github.com/oracle/oci-manager/pkg/controller/oci/resources/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

var (
	vcntest1 = corev1alpha1.Vcn{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vcn.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       corev1alpha1.VirtualNetworkKind,
		},
		Spec: corev1alpha1.VcnSpec{
			DisplayName:    "testDisplay",
			CompartmentRef: "compartment.test1",
		},
		Status: corev1alpha1.VcnStatus{
			Resource: &corev1alpha1.VcnResource{
				Vcn: ocisdkcore.Vcn{Id: resourcescommon.StrPtrOrNil("fakeVcnOCIID")},
			},
		},
	}
	fakeImageMap = map[string]string{
		"Oracle-Linux-7.4-2018.02.21-1": "testImage",
	}
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
			Images: fakeImageMap,
			Resource: &identityv1alpha1.CompartmentResource{
				Compartment: ocisdkidentity.Compartment{Id: resourcescommon.StrPtrOrNil("fakeCompartmentOCIID")},
			},
		},
	}
	subnet = corev1alpha1.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "subnet.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       corev1alpha1.SubnetKind,
		},
		Spec: corev1alpha1.SubnetSpec{
			DisplayName: "testCompartment"},
		Status: corev1alpha1.SubnetStatus{
			Resource: &corev1alpha1.SubnetResource{
				Subnet: ocisdkcore.Subnet{Id: resourcescommon.StrPtrOrNil("fakeSubnetOCIID")},
			},
		},
	}
)

func TestVcnResourceBasic(t *testing.T) {

	t.Log("Testing vcn_resource")
	clientset := fakeclient.NewSimpleClientset()
	vcnClient := fakeoci.NewVcnClient()

	vcnAdapter := VcnAdapter{}
	vcnAdapter.clientset = clientset
	vcnAdapter.vcnClient = vcnClient

	comp, err := clientset.OciidentityV1alpha1().Compartments(fakeNs).Create(&compartment)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created compartment object %v", comp)

	newVcn, err := vcnAdapter.CreateObject(&vcntest1)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Vcn object %v", newVcn)

	if vcnAdapter.IsExpectedType(newVcn) {
		t.Logf("Checked Vcn type")
	}

	vcnWithResource, err := vcnAdapter.Create(newVcn)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Vcn resource %v", vcnWithResource)

	vcnWithResource, err = vcnAdapter.Get(newVcn)
	vcnWithResource, err = vcnAdapter.Update(newVcn)
	vcnWithResource, err = vcnAdapter.Delete(newVcn)
	eq := vcnAdapter.Equivalent(newVcn, newVcn)
	if !eq {
		t.Errorf("should be equal")
	}
	_, err = vcnAdapter.DependsOnRefs(newVcn)
}

func NewFakeVcnAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	vcnClient := fakeoci.NewVcnClient()
	vcnAdapter := VcnAdapter{}
	vcnAdapter.clientset = clientset
	vcnAdapter.vcnClient = vcnClient
	return &vcnAdapter
}
