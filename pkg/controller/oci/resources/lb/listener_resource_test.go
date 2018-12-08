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

	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	fakeoci "github.com/oracle/oci-manager/pkg/controller/oci/resources/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// ocisdkcore "github.com/oracle/oci-go-sdk/core"
	ocisdklb "github.com/oracle/oci-go-sdk/loadbalancer"
	// ocisdkidentity "github.com/oracle/oci-go-sdk/identity"

	lbv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocilb.oracle.com/v1alpha1"
	fakeclient "github.com/oracle/oci-manager/pkg/client/clientset/versioned/fake"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
)

var (
	lbTest1 = lbv1alpha1.LoadBalancer{
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
				LoadBalancer: &ocisdklb.LoadBalancer{Id: resourcescommon.StrPtrOrNil("fakeLoadBalancerOCIID")},
			},
		},
	}

	bsTest1 = lbv1alpha1.BackendSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "backendSet.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       lbv1alpha1.BackendSetKind,
		},
		Spec: lbv1alpha1.BackendSetSpec{
			Policy: "ROUND_ROBIN",
		},
		Status: lbv1alpha1.BackendSetStatus{
			Resource: &lbv1alpha1.BackendSetResource{
				BackendSet: &ocisdklb.BackendSet{Name: resourcescommon.StrPtrOrNil("fakeBackendSet")},
			},
		},
	}

	listenertest1 = lbv1alpha1.Listener{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "listener.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       lbv1alpha1.ListenerKind,
		},
		Spec: lbv1alpha1.ListenerSpec{
			LoadBalancerRef:       "loadBalancer.test1",
			Protocol:              "test-proto",
			DefaultBackendSetName: "backendSet.test1",
			Port:                  1983,
		},
	}
)

func TestListenerResourceBasic(t *testing.T) {

	t.Log("Testing listener_resource")
	clientset := fakeclient.NewSimpleClientset()
	lbClient := fakeoci.NewLoadBalancerClient()

	listenerAdapter := ListenerAdapter{}
	listenerAdapter.clientset = clientset
	listenerAdapter.lbClient = lbClient

	lb, err := clientset.OcilbV1alpha1().LoadBalancers(fakeNs).Create(&lbTest1)
	if err != nil {
		t.Errorf("Got lb error %v", err)
	}
	t.Logf("Created lb object %v", lb)

	bs, err := clientset.OcilbV1alpha1().BackendSets(fakeNs).Create(&bsTest1)
	if err != nil {
		t.Errorf("Got bs error %v", err)
	}
	t.Logf("Created bs object %v", bs)

	newListener, err := listenerAdapter.CreateObject(&listenertest1)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Listener object %v", newListener)

	if listenerAdapter.IsExpectedType(newListener) {
		t.Logf("Checked Listener type")
	}

	listenerWithResource, err := listenerAdapter.Create(newListener)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Listener resource %v", listenerWithResource)

	listenerWithResource, err = listenerAdapter.Get(listenerWithResource)
	_, err = listenerAdapter.Update(listenerWithResource)
	_, err = listenerAdapter.Delete(listenerWithResource)
	_, err = listenerAdapter.DependsOnRefs(listenerWithResource)

}

func NewFakeListenerAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	listenerClient := fakeoci.NewLoadBalancerClient()
	listenerAdapter := ListenerAdapter{}
	listenerAdapter.clientset = clientset
	listenerAdapter.lbClient = listenerClient
	return &listenerAdapter
}
