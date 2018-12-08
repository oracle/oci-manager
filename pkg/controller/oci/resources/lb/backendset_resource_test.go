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

	lbv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocilb.oracle.com/v1alpha1"
	fakeclient "github.com/oracle/oci-manager/pkg/client/clientset/versioned/fake"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
)

var (
	proto   = "HTTP"
	port    = 1983
	urlPath = "/"

	hc = &lbv1alpha1.HealthChecker{
		Protocol: proto,
		Port:     port,
		URLPath:  urlPath,
	}

	backendSettest1 = lbv1alpha1.BackendSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "backendSet.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       lbv1alpha1.BackendSetKind,
		},
		Spec: lbv1alpha1.BackendSetSpec{
			HealthChecker:   hc,
			LoadBalancerRef: "loadBalancer.test1",
			Policy:          "ROUND_ROBIN",
		},
	}
)

func TestBackendSetResourceBasic(t *testing.T) {

	t.Log("Testing backendSet_resource")
	clientset := fakeclient.NewSimpleClientset()
	lbClient := fakeoci.NewLoadBalancerClient()

	backendSetAdapter := BackendSetAdapter{}
	backendSetAdapter.clientset = clientset
	backendSetAdapter.lbClient = lbClient

	lb, err := clientset.OcilbV1alpha1().LoadBalancers(fakeNs).Create(&lbTest1)
	if err != nil {
		t.Errorf("Got lb error %v", err)
	}
	t.Logf("Created lb object %v", lb)

	newBackendSet, err := backendSetAdapter.CreateObject(&backendSettest1)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created BackendSet object %v", newBackendSet)

	if backendSetAdapter.IsExpectedType(newBackendSet) {
		t.Logf("Checked BackendSet type")
	}

	backendSetWithResource, err := backendSetAdapter.Create(newBackendSet)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created BackendSet resource %v", backendSetWithResource)

	backendSetWithResource, err = backendSetAdapter.Get(backendSetWithResource)
	_, err = backendSetAdapter.Update(backendSetWithResource)
	_, err = backendSetAdapter.Delete(backendSetWithResource)
	_, err = backendSetAdapter.DependsOnRefs(newBackendSet)
}

func NewFakeBackendSetAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	backendSetClient := fakeoci.NewLoadBalancerClient()
	backendSetAdapter := BackendSetAdapter{}
	backendSetAdapter.clientset = clientset
	backendSetAdapter.lbClient = backendSetClient
	return &backendSetAdapter
}
