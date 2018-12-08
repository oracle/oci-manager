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

package ce

import (
	cev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocice.oracle.com/v1alpha1"
	corev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	identityv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com/v1alpha1"
	fakeclient "github.com/oracle/oci-manager/pkg/client/clientset/versioned/fake"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"

	ocisdkce "github.com/oracle/oci-go-sdk/containerengine"
	ocisdkcore "github.com/oracle/oci-go-sdk/core"
	ocisdkidentity "github.com/oracle/oci-go-sdk/identity"

	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	fakeoci "github.com/oracle/oci-manager/pkg/controller/oci/resources/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

var (
	fakeNs   = "fakens"
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
				Vcn: ocisdkcore.Vcn{Id: resourcescommon.StrPtrOrNil("fakeClusterOCIID")},
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
	subnet1 = corev1alpha1.Subnet{
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

	subnet2 = corev1alpha1.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "subnet.test2",
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

	k8sVersion       = "v1.9.7"
	serviceLbSubnets = []string{"subnet.test1", "subnet.test2"}
	podsCidr         = "10.97.0.0/16"
	serviceCidr      = "10.96.0.0/16"

	options = ocisdkce.ClusterCreateOptions{
		KubernetesNetworkConfig: &ocisdkce.KubernetesNetworkConfig{
			PodsCidr:     &podsCidr,
			ServicesCidr: &serviceCidr,
		},
	}

	fakeClusterId = "1"

	clustertest1 = cev1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocice.oracle.com/v1alpha1",
			Kind:       cev1alpha1.ClusterKind,
		},
		Spec: cev1alpha1.ClusterSpec{
			CompartmentRef:      "compartment.test1",
			VcnRef:              "vcn.test1",
			KubernetesVersion:   &k8sVersion,
			ServiceLbSubnetRefs: serviceLbSubnets,
			Options:             &options,
		},
		Status: cev1alpha1.ClusterStatus{
			Resource: &cev1alpha1.ClusterResource{
				Cluster: &ocisdkce.Cluster{
					Id: &fakeClusterId,
				},
			},
		},
	}
)

func TestClusterResourceBasic(t *testing.T) {

	t.Log("Testing cluster_resource")
	clientset := fakeclient.NewSimpleClientset()
	clusterAdapter := NewFakeClusterAdapter(clientset)

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

	sn1, err := clientset.OcicoreV1alpha1().Subnets(fakeNs).Create(&subnet1)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created subnet1 object %v", sn1)

	sn2, err := clientset.OcicoreV1alpha1().Subnets(fakeNs).Create(&subnet2)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created subnet2 object %v", sn2)

	newCluster, err := clusterAdapter.CreateObject(&clustertest1)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Cluster object %v", newCluster)

	if clusterAdapter.IsExpectedType(newCluster) {
		t.Logf("Checked Cluster type")
	}

	clusterWithResource, err := clusterAdapter.Create(newCluster)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Cluster resource %v", clusterWithResource)

	clusterWithResource, err = clusterAdapter.Get(newCluster)
	t.Logf("Get Cluster resource %v", clusterWithResource)

	_, err = clusterAdapter.Update(clusterWithResource)
	t.Logf("Update Cluster resource %v", clusterWithResource)

	_, err = clusterAdapter.Delete(clusterWithResource)
	t.Logf("Delete Cluster resource %v", clusterWithResource)

	_, err = clusterAdapter.DependsOnRefs(newCluster)
}

func NewFakeClusterAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	clusterClient := fakeoci.NewContainerEngineClient()
	clusterAdapter := ClusterAdapter{}
	clusterAdapter.clientset = clientset
	clusterAdapter.ceClient = clusterClient
	return &clusterAdapter
}
