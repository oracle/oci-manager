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

	corev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	fakeclient "github.com/oracle/oci-manager/pkg/client/clientset/versioned/fake"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
)

const fakeNs = "fakeNs"

var (
	volumetest1 = corev1alpha1.Volume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "volume.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       corev1alpha1.VolumeKind,
		},
		Spec: corev1alpha1.VolumeSpec{
			DisplayName:    "bla",
			AttachmentType: "ISCSI",
		},
	}
)

func TestVolumeResourceBasic(t *testing.T) {

	t.Log("Testing volume_resource")
	clientset := fakeclient.NewSimpleClientset()
	bsClient := fakeoci.NewBlockStorageClient()

	volumeAdapter := VolumeAdapter{}
	volumeAdapter.clientset = clientset
	volumeAdapter.bsClient = bsClient

	comp, err := clientset.OciidentityV1alpha1().Compartments(fakeNs).Create(&compartment)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created compartment object %v", comp)

	newVolume, err := volumeAdapter.CreateObject(&volumetest1)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Volume object %v", newVolume)

	if volumeAdapter.IsExpectedType(newVolume) {
		t.Logf("Checked Volume type")
	}

	volumeWithResource, err := volumeAdapter.Create(newVolume)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Volume resource %v", volumeWithResource)

	volumeWithResource, err = volumeAdapter.Get(newVolume)
	volumeWithResource, err = volumeAdapter.Update(newVolume)
	volumeWithResource, err = volumeAdapter.Delete(newVolume)
	eq := volumeAdapter.Equivalent(newVolume, newVolume)
	if !eq {
		t.Errorf("should be equal")
	}
	_, err = volumeAdapter.DependsOnRefs(newVolume)
}

func NewFakeVolumeAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	bsClient := fakeoci.NewBlockStorageClient()
	volumeAdapter := VolumeAdapter{}
	volumeAdapter.clientset = clientset
	volumeAdapter.bsClient = bsClient
	return &volumeAdapter
}
