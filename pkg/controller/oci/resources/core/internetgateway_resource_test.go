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
	internetGatewaytest1 = corev1alpha1.InternetGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "internetGateway.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       corev1alpha1.InternetGatewayKind,
		},
		Spec: corev1alpha1.InternetGatewaySpec{
			CompartmentRef: "compartment.test1",
			VcnRef:         "vcn.test1",
			DisplayName:    "bla",
		},
	}
)

func TestInternetGatewayResourceBasic(t *testing.T) {

	t.Log("Testing internetGateway_resource")
	clientset := fakeclient.NewSimpleClientset()
	vcnClient := fakeoci.NewVcnClient()

	internetGatewayAdapter := InternetGatewayAdapter{}
	internetGatewayAdapter.clientset = clientset
	internetGatewayAdapter.vcnClient = vcnClient

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

	newInternetGateway, err := internetGatewayAdapter.CreateObject(&internetGatewaytest1)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created InternetGateway object %v", newInternetGateway)

	if internetGatewayAdapter.IsExpectedType(newInternetGateway) {
		t.Logf("Checked InternetGateway type")
	}

	internetGatewayWithResource, err := internetGatewayAdapter.Create(newInternetGateway)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created InternetGateway resource %v", internetGatewayWithResource)

	internetGatewayWithResource, err = internetGatewayAdapter.Get(newInternetGateway)
	internetGatewayWithResource, err = internetGatewayAdapter.Update(newInternetGateway)
	internetGatewayWithResource, err = internetGatewayAdapter.Delete(newInternetGateway)
	eq := internetGatewayAdapter.Equivalent(newInternetGateway, newInternetGateway)
	if !eq {
		t.Errorf("should be equal")
	}
	_, err = internetGatewayAdapter.DependsOnRefs(newInternetGateway)
}

func NewFakeInternetGatewayAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	vcnClient := fakeoci.NewVcnClient()
	internetGatewayAdapter := InternetGatewayAdapter{}
	internetGatewayAdapter.clientset = clientset
	internetGatewayAdapter.vcnClient = vcnClient
	return &internetGatewayAdapter
}
