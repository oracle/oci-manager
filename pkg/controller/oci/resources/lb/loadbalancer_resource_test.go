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

package lb

import (
	"testing"

	ocisdkidentity "github.com/oracle/oci-go-sdk/identity"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	fakeoci "github.com/oracle/oci-manager/pkg/controller/oci/resources/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ocisdkcore "github.com/oracle/oci-go-sdk/core"
	ocisdklb "github.com/oracle/oci-go-sdk/loadbalancer"

	corev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	identityv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com/v1alpha1"
	lbv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocilb.oracle.com/v1alpha1"
	fakeclient "github.com/oracle/oci-manager/pkg/client/clientset/versioned/fake"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
)

const fakeNs = "fakeNs"

var (
	testid = "foo"

	loadBalancertest1 = lbv1alpha1.LoadBalancer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "loadBalancer.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       lbv1alpha1.LoadBalancerKind,
		},
		Spec: lbv1alpha1.LoadBalancerSpec{
			CompartmentRef: "compartment.test1",
			SubnetRefs: []string{
				"subnet.test1",
			},
		},
		Status: lbv1alpha1.LoadBalancerStatus{
			Resource: &lbv1alpha1.LoadBalancerResource{
				LoadBalancer: &ocisdklb.LoadBalancer{
					Id: &testid,
				},
			},
		},
	}

	images = map[string]string{
		"Oracle-Linux-7.4-2018.02.21-1": "testImage",
	}

	compartment = identityv1alpha1.Compartment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "compartment.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ociidentity.oracle.com/v1alpha1",
			Kind:       identityv1alpha1.CompartmentKind,
		},
		Spec: identityv1alpha1.CompartmentSpec{
			Description: "testCompartment"},
		Status: identityv1alpha1.CompartmentStatus{
			Images: images,
			Resource: &identityv1alpha1.CompartmentResource{
				Compartment: ocisdkidentity.Compartment{
					Id: resourcescommon.StrPtrOrNil("fakeCompartmentOCIID"),
				},
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

func TestLoadBalancerResourceBasic(t *testing.T) {

	t.Log("Testing loadBalancer_resource")
	clientset := fakeclient.NewSimpleClientset()
	loadBalancerClient := fakeoci.NewLoadBalancerClient()

	loadBalancerAdapter := LoadBalancerAdapter{}
	loadBalancerAdapter.clientset = clientset
	loadBalancerAdapter.lbClient = loadBalancerClient

	comp, err := clientset.OciidentityV1alpha1().Compartments(fakeNs).Create(&compartment)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created compartment object %v", comp)

	subnet, err := clientset.OcicoreV1alpha1().Subnets(fakeNs).Create(&subnet)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created subnet object %v", subnet)

	newLoadBalancer, err := loadBalancerAdapter.CreateObject(&loadBalancertest1)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created LoadBalancer object %v", newLoadBalancer)

	if loadBalancerAdapter.IsExpectedType(newLoadBalancer) {
		t.Logf("Checked LoadBalancer type")
	}

	loadBalancerWithResource, err := loadBalancerAdapter.Create(newLoadBalancer)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created LoadBalancer resource %v", loadBalancerWithResource)

	loadBalancerWithResource, err = loadBalancerAdapter.Get(loadBalancerWithResource)

	_, err = loadBalancerAdapter.Delete(loadBalancerWithResource)
	_, err = loadBalancerAdapter.DependsOnRefs(loadBalancerWithResource)
}

func NewFakeLoadBalancerAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	loadBalancerClient := fakeoci.NewLoadBalancerClient()
	loadBalancerAdapter := LoadBalancerAdapter{}
	loadBalancerAdapter.clientset = clientset
	loadBalancerAdapter.lbClient = loadBalancerClient
	return &loadBalancerAdapter
}
