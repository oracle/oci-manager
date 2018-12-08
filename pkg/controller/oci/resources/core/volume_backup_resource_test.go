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
	volumeBackuptest1 = corev1alpha1.VolumeBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "volumeBackup.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       corev1alpha1.VolumeBackupKind,
		},
		Spec: corev1alpha1.VolumeBackupSpec{
			VolumeRef: "volume.test1",

			DisplayName:      "bla",
			VolumeBackupType: "FULL",
		},
	}

	volume = corev1alpha1.Volume{
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
		Status: corev1alpha1.VolumeStatus{
			Resource: &corev1alpha1.VolumeResource{
				Volume: ocisdkcore.Volume{
					Id: resourcescommon.StrPtrOrNil("fakeVolumeOCIID"),
				},
			},
		},
	}
)

func TestVolumeBackupResourceBasic(t *testing.T) {

	t.Log("Testing VolumeBackup_resource")
	clientset := fakeclient.NewSimpleClientset()
	bsClient := fakeoci.NewBlockStorageClient()

	volumeBackupAdapter := VolumeBackupAdapter{}
	volumeBackupAdapter.clientset = clientset
	volumeBackupAdapter.bsClient = bsClient

	vol, err := clientset.OcicoreV1alpha1().Volumes(fakeNs).Create(&volume)
	if err != nil {
		t.Errorf("Got volume error %v", err)
	}
	t.Logf("Created volume object %v", vol)

	newVolumeBackup, err := volumeBackupAdapter.CreateObject(&volumeBackuptest1)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created VolumeBackup object %v", newVolumeBackup)

	if volumeBackupAdapter.IsExpectedType(newVolumeBackup) {
		t.Logf("Checked VolumeBackup type")
	}
	volumeBackupWithResource, err := volumeBackupAdapter.Create(newVolumeBackup)

	if err != nil {
		t.Errorf("Got create volume error %v", err)
	}
	t.Logf("Created VolumeBackup resource %v", volumeBackupWithResource)

	volumeBackupWithResource, err = volumeBackupAdapter.Get(newVolumeBackup)
	volumeBackupWithResource, err = volumeBackupAdapter.Update(newVolumeBackup)
	volumeBackupWithResource, err = volumeBackupAdapter.Delete(newVolumeBackup)
	eq := volumeBackupAdapter.Equivalent(newVolumeBackup, newVolumeBackup)
	if !eq {
		t.Errorf("should be equal")
	}
	_, err = volumeBackupAdapter.DependsOnRefs(newVolumeBackup)
}

func NewFakevolumeBackupAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	bsClient := fakeoci.NewBlockStorageClient()
	volumeBackupAdapter := VolumeBackupAdapter{}
	volumeBackupAdapter.clientset = clientset
	volumeBackupAdapter.bsClient = bsClient
	return &volumeBackupAdapter
}
