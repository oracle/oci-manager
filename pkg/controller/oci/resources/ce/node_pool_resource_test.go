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
	fakeclient "github.com/oracle/oci-manager/pkg/client/clientset/versioned/fake"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"

	ocisdkce "github.com/oracle/oci-go-sdk/containerengine"

	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	fakeoci "github.com/oracle/oci-manager/pkg/controller/oci/resources/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

var (
	one       = 1
	nodeShape = "vm1.1"
	okeImage  = "Oracle-Linux-7.4"
	subnets   = []string{"subnet.test1", "subnet.test2"}

	nodepooltest1 = cev1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nodepool.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocice.oracle.com/v1alpha1",
			Kind:       cev1alpha1.NodePoolKind,
		},
		Spec: cev1alpha1.NodePoolSpec{
			CompartmentRef:    "compartment.test1",
			ClusterRef:        "cluster.test1",
			QuantityPerSubnet: &one,
			NodeImageName:     &okeImage,
			KubernetesVersion: &k8sVersion,
			NodeShape:         &nodeShape,
			SubnetRefs:        subnets,
		},
		Status: cev1alpha1.NodePoolStatus{
			Resource: &cev1alpha1.NodePoolResource{
				NodePool: &ocisdkce.NodePool{
					Id: &fakeClusterId,
				},
			},
		},
	}
)

func TestNodePoolResourceBasic(t *testing.T) {

	t.Log("Testing nodepool_resource")
	clientset := fakeclient.NewSimpleClientset()
	nodepoolAdapter := NewFakeNodePoolAdapter(clientset)

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

	cluster, err := clientset.OciceV1alpha1().Clusters(fakeNs).Create(&clustertest1)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created cluster object %v", cluster)

	newNodepool, err := nodepoolAdapter.CreateObject(&nodepooltest1)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Nodepool object %v", newNodepool)

	if nodepoolAdapter.IsExpectedType(newNodepool) {
		t.Logf("Checked Nodepool type")
	}

	nodepoolWithResource, err := nodepoolAdapter.Create(newNodepool)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Nodepool resource %v", nodepoolWithResource)

	_, err = nodepoolAdapter.Get(newNodepool)
	t.Logf("Get Nodepool resource %v", nodepoolWithResource)

	_, err = nodepoolAdapter.Update(nodepoolWithResource)
	t.Logf("Update Nodepool resource %v", nodepoolWithResource)

	_, err = nodepoolAdapter.Delete(nodepoolWithResource)
	t.Logf("Delete Nodepool resource %v", nodepoolWithResource)

	_, err = nodepoolAdapter.DependsOnRefs(nodepoolWithResource)
}

func NewFakeNodePoolAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	nodepoolClient := fakeoci.NewContainerEngineClient()
	nodepoolAdapter := NodePoolAdapter{}
	nodepoolAdapter.clientset = clientset
	nodepoolAdapter.ceClient = nodepoolClient
	return &nodepoolAdapter
}
