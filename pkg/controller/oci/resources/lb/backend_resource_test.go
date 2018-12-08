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

	ocisdkcore "github.com/oracle/oci-go-sdk/core"

	corev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	lbv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocilb.oracle.com/v1alpha1"
	fakeclient "github.com/oracle/oci-manager/pkg/client/clientset/versioned/fake"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
)

var (
	ip = "1.1.1.1"

	instance = corev1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       corev1alpha1.InstanceKind,
		},
		Spec: corev1alpha1.InstanceSpec{
			CompartmentRef: "compartment.test1",
			SubnetRef:      "subnet.test1",

			DisplayName:        "bla",
			Image:              "Oracle-Linux-7.4-2018.02.21-1",
			AvailabilityDomain: "ad.test1",
			Shape:              "shape.test1",
			Metadata:           make(map[string]string),
			ExtendedMetadata:   make(map[string]interface{}),
			HostnameLabel:      "test",
			IpxeScript:         "",
		},
		Status: corev1alpha1.InstanceStatus{
			Resource: &corev1alpha1.InstanceResource{
				Instance: ocisdkcore.Instance{
					Id: resourcescommon.StrPtrOrNil("fakeInstanceOCIID"),
				},
			},
			PrimaryVnic: &corev1alpha1.PrimaryVnicResource{
				Vnic: ocisdkcore.Vnic{
					PrivateIp: &ip,
				},
			},
		},
	}

	backendtest1 = lbv1alpha1.Backend{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "backend.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       lbv1alpha1.BackendKind,
		},
		Spec: lbv1alpha1.BackendSpec{
			InstanceRef: "instance.test1",
			IPAddress:   "1.1.1.1",
			Port:        1983,
			Weight:      1,
		},
	}
)

func TestBackendResourceBasic(t *testing.T) {

	t.Log("Testing backend_resource")
	clientset := fakeclient.NewSimpleClientset()

	cClient := fakeoci.NewComputeClient()
	lbClient := fakeoci.NewLoadBalancerClient()
	vcnClient := fakeoci.NewVcnClient()

	backendAdapter := BackendAdapter{}
	backendAdapter.clientset = clientset
	backendAdapter.lbClient = lbClient
	backendAdapter.cClient = cClient
	backendAdapter.vcnClient = vcnClient

	lb, err := clientset.OcilbV1alpha1().LoadBalancers(fakeNs).Create(&lbTest1)
	if err != nil {
		t.Errorf("Got lb error %v", err)
	}
	t.Logf("Created lb object %v", lb)

	i, err := clientset.OcicoreV1alpha1().Instances(fakeNs).Create(&instance)
	if err != nil {
		t.Errorf("Got instance error %v", err)
	}
	t.Logf("Created instance object %v", i)

	newBackend, err := backendAdapter.CreateObject(&backendtest1)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Backend object %v", newBackend)

	if backendAdapter.IsExpectedType(newBackend) {
		t.Logf("Checked Backend type")
	}

	backendWithResource, err := backendAdapter.Create(newBackend)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Backend resource %v", backendWithResource)

	backendWithResource, err = backendAdapter.Get(backendWithResource)
	_, err = backendAdapter.Update(backendWithResource)
	_, err = backendAdapter.Delete(backendWithResource)
	_, err = backendAdapter.DependsOnRefs(backendWithResource)
}

func NewFakeBackendAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	backendClient := fakeoci.NewLoadBalancerClient()
	backendAdapter := BackendAdapter{}
	backendAdapter.clientset = clientset
	backendAdapter.lbClient = backendClient
	return &backendAdapter
}
