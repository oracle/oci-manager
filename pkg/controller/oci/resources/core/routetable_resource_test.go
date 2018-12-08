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

	ocisdkcore "github.com/oracle/oci-go-sdk/core"

	corev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	fakeclient "github.com/oracle/oci-manager/pkg/client/clientset/versioned/fake"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
)

var (
	routeTabletest1 = corev1alpha1.RouteTable{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "routeTable.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       corev1alpha1.RouteTableKind,
		},
		Spec: corev1alpha1.RouteTableSpec{
			DisplayName: "bla",
		},
		Status: corev1alpha1.RouteTableStatus{
			Resource: &corev1alpha1.RouteTableResource{
				RouteTable: ocisdkcore.RouteTable{Id: resourcescommon.StrPtrOrNil("fakeRouteTableOCIID")},
			},
		},
	}
)

func TestRouteTableResourceBasic(t *testing.T) {

	t.Log("Testing routeTable_resource")
	clientset := fakeclient.NewSimpleClientset()
	vcnClient := fakeoci.NewVcnClient()

	routeTableAdapter := RouteTableAdapter{}
	routeTableAdapter.clientset = clientset
	routeTableAdapter.vcnClient = vcnClient

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

	newRouteTable, err := routeTableAdapter.CreateObject(&routeTabletest1)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created RouteTable object %v", newRouteTable)

	if routeTableAdapter.IsExpectedType(newRouteTable) {
		t.Logf("Checked RouteTable type")
	}

	routeTableWithResource, err := routeTableAdapter.Create(newRouteTable)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created RouteTable resource %v", routeTableWithResource)

	routeTableWithResource, err = routeTableAdapter.Get(newRouteTable)
	routeTableWithResource, err = routeTableAdapter.Update(newRouteTable)
	routeTableWithResource, err = routeTableAdapter.Delete(newRouteTable)
	eq := routeTableAdapter.Equivalent(newRouteTable, newRouteTable)
	if !eq {
		t.Errorf("should be equal")
	}
	_, err = routeTableAdapter.DependsOnRefs(newRouteTable)
}

func NewFakeRouteTableAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	vcnClient := fakeoci.NewVcnClient()
	routeTableAdapter := RouteTableAdapter{}
	routeTableAdapter.clientset = clientset
	routeTableAdapter.vcnClient = vcnClient
	return &routeTableAdapter
}
