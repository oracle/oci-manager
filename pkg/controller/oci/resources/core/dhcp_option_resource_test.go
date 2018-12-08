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
	dhcpOptions = corev1alpha1.DhcpOption{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dhcpOption.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       corev1alpha1.DhcpOptionKind,
		},
		Spec: corev1alpha1.DhcpOptionSpec{
			CompartmentRef: "compartment.test1",
			VcnRef:         "vcn.test1",
			DisplayName:    "bla",
		},
	}
)

func TestDhcpOptionResourceBasic(t *testing.T) {

	t.Log("Testing dhcpOption_resource")
	clientset := fakeclient.NewSimpleClientset()
	vcnClient := fakeoci.NewVcnClient()

	dhcpOptionAdapter := DhcpOptionAdapter{}
	dhcpOptionAdapter.clientset = clientset
	dhcpOptionAdapter.vcnClient = vcnClient

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

	options := []ocisdkcore.DhcpOption{}
	option := ocisdkcore.DhcpDnsOption{
		CustomDnsServers: []string{"127.0.0.1"},
	}
	options = append(options, option)
	dhcpOptions.Spec.Options = options

	newDhcpOption, err := dhcpOptionAdapter.CreateObject(&dhcpOptions)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created DhcpOption object %v", newDhcpOption)

	if dhcpOptionAdapter.IsExpectedType(newDhcpOption) {
		t.Logf("Checked DhcpOption type")
	}

	dhcpOptionWithResource, err := dhcpOptionAdapter.Create(newDhcpOption)

	if err != nil {
		t.Errorf("Got create dhcp option error %v", err)
	}
	t.Logf("Created DhcpOption resource %v", dhcpOptionWithResource)

	dhcpOptionWithResource, err = dhcpOptionAdapter.Get(newDhcpOption)
	dhcpOptionWithResource, err = dhcpOptionAdapter.Update(newDhcpOption)
	dhcpOptionWithResource, err = dhcpOptionAdapter.Delete(newDhcpOption)
	eq := dhcpOptionAdapter.Equivalent(newDhcpOption, newDhcpOption)
	if !eq {
		t.Errorf("should be equal")
	}
	_, err = dhcpOptionAdapter.DependsOnRefs(newDhcpOption)
}

func NewFakeDhcpOptionAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	vcnClient := fakeoci.NewVcnClient()
	dhcpOptionAdapter := DhcpOptionAdapter{}
	dhcpOptionAdapter.clientset = clientset
	dhcpOptionAdapter.vcnClient = vcnClient
	return &dhcpOptionAdapter
}
