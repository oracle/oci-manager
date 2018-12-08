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

	// ocisdkidentity "github.com/oracle/oci-go-sdk/identity"

	identityv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com/v1alpha1"
	fakeclient "github.com/oracle/oci-manager/pkg/client/clientset/versioned/fake"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
)

const fakeNs = "fakeNs"

var (
	compartment1 = identityv1alpha1.Compartment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "compartment.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       identityv1alpha1.CompartmentKind,
		},
		Spec: identityv1alpha1.CompartmentSpec{
			Description: "testCompartment",
		},
	}
)

func TestCompartmentResourceBasic(t *testing.T) {

	t.Log("Testing compartment_resource")
	clientset := fakeclient.NewSimpleClientset()
	idClient := fakeoci.NewIdentityClient()
	cClient := fakeoci.NewComputeClient()

	compartmentAdapter := CompartmentAdapter{}
	compartmentAdapter.clientset = clientset
	compartmentAdapter.ociIdClient = idClient
	compartmentAdapter.ociCoreComputeClient = cClient

	newCompartment, err := compartmentAdapter.CreateObject(&compartment1)
	if err != nil {
		t.Errorf("Got compartment CreateObject error %v", err)
	}
	t.Logf("Created Compartment object %v", newCompartment)

	if compartmentAdapter.IsExpectedType(newCompartment) {
		t.Logf("Checked Compartment type")
	}

	compartmentWithResource, err := compartmentAdapter.Create(newCompartment)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Get Compartment resource %v", compartmentWithResource)

	compartmentWithResource, err = compartmentAdapter.Get(newCompartment)
	compartmentWithResource, err = compartmentAdapter.Update(newCompartment)
	compartmentWithResource, err = compartmentAdapter.Delete(newCompartment)
	_, err = compartmentAdapter.DependsOnRefs(newCompartment)

}

func NewFakeCompartmentAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	idClient := fakeoci.NewIdentityClient()
	compartmentAdapter := CompartmentAdapter{}
	compartmentAdapter.clientset = clientset
	compartmentAdapter.ociIdClient = idClient
	return &compartmentAdapter
}
