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
	ocisdkidentity "github.com/oracle/oci-go-sdk/identity"

	corev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	identityv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com/v1alpha1"

	fakeclient "github.com/oracle/oci-manager/pkg/client/clientset/versioned/fake"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
)

var (
	instancetest1 = corev1alpha1.Instance{
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
	}

	instancetest2 = corev1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance.test2",
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
					Id:            resourcescommon.StrPtrOrNil("fakeInstanceOCIID"),
					CompartmentId: resourcescommon.StrPtrOrNil("fakeInstanceOCIID"),
				},
			},
			PrimaryVnic: &corev1alpha1.PrimaryVnicResource{
				Vnic: ocisdkcore.Vnic{
					PrivateIp: resourcescommon.StrPtrOrNil("1.1.1.1"),
				},
			},
		},
	}

	fakeImageMapForInstance = map[string]string{
		"Oracle-Linux-7.4-2018.02.21-1": "testImage",
	}

	compartmentForInstance = identityv1alpha1.Compartment{
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
			Images: fakeImageMapForInstance,
			Resource: &identityv1alpha1.CompartmentResource{
				Compartment: ocisdkidentity.Compartment{Id: resourcescommon.StrPtrOrNil("fakeCompartmentOCIID")},
			},
		},
	}

	subnetForInstance = corev1alpha1.Subnet{
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

func TestInstanceResourceBasic(t *testing.T) {

	t.Log("Testing instance_resource")
	clientset := fakeclient.NewSimpleClientset()
	vcnClient := fakeoci.NewVcnClient()
	ccClient := fakeoci.NewComputeClient()
	bsClient := fakeoci.NewBlockStorageClient()

	instanceAdapter := InstanceAdapter{}
	instanceAdapter.clientset = clientset
	instanceAdapter.vcnClient = vcnClient
	instanceAdapter.cClient = ccClient
	instanceAdapter.bsClient = bsClient

	comp, err := clientset.OciidentityV1alpha1().Compartments(fakeNs).Create(&compartmentForInstance)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created compartment object %v", comp)

	subnet, err := clientset.OcicoreV1alpha1().Subnets(fakeNs).Create(&subnetForInstance)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created subnet object %v", subnet)

	newInstance, err := instanceAdapter.CreateObject(&instancetest1)

	if err != nil {
		t.Errorf("Got instance CreateObject error %v", err)
	}
	t.Logf("Created Instance object %v", newInstance)

	if instanceAdapter.IsExpectedType(newInstance) {
		t.Logf("Checked Instance type")
	}

	instanceWithResource, err := instanceAdapter.Create(newInstance)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Instance resource %v", instanceWithResource)

	instanceWithResource, err = instanceAdapter.Get(newInstance)

	i, err := clientset.OcicoreV1alpha1().Instances(fakeNs).Create(&instancetest2)
	if err != nil {
		t.Errorf("Got instance error %v", err)
	}
	t.Logf("Created instance object %v", i)

	instanceWithResource, err = instanceAdapter.Update(newInstance)
	instanceWithResource, err = instanceAdapter.Delete(newInstance)
	eq := instanceAdapter.Equivalent(newInstance, newInstance)
	if !eq {
		t.Errorf("should be equal")
	}
	_, err = instanceAdapter.DependsOnRefs(newInstance)
}

func NewFakeInstanceAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	vcnClient := fakeoci.NewVcnClient()
	instanceAdapter := InstanceAdapter{}
	instanceAdapter.clientset = clientset
	instanceAdapter.vcnClient = vcnClient
	return &instanceAdapter
}
