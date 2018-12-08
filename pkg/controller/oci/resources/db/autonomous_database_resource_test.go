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

package db

import (
	"math/rand"
	"testing"
	"time"

	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	fakeoci "github.com/oracle/oci-manager/pkg/controller/oci/resources/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ocisdkdb "github.com/oracle/oci-go-sdk/database"
	ocisdkidentity "github.com/oracle/oci-go-sdk/identity"

	dbv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocidb.oracle.com/v1alpha1"
	identityv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com/v1alpha1"

	fakeclient "github.com/oracle/oci-manager/pkg/client/clientset/versioned/fake"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"

	fakekube "k8s.io/client-go/kubernetes/fake"
)

const (
	fakeNs = "fakens"
)

var (
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

	one    = 1
	fakeId = "123"

	adb = dbv1alpha1.AutonomousDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name: "adb.test1",
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocidb.oracle.com/v1alpha1",
			Kind:       dbv1alpha1.AutonomousDatabaseKind,
		},
		Spec: dbv1alpha1.AutonomousDatabaseSpec{
			CompartmentRef:       "compartment.test1",
			DisplayName:          "bla",
			CpuCoreCount:         &one,
			DataStorageSizeInTBs: &one,
		},
		Status: dbv1alpha1.AutonomousDatabaseStatus{
			Resource: &dbv1alpha1.AutonomousDatabaseResource{
				AutonomousDatabase: &ocisdkdb.AutonomousDatabase{
					Id: &fakeId,
				},
			},
		},
	}
)

func TestAutonomousDatabase(t *testing.T) {

	t.Log("Testing AutonomousDatabase resource")
	clientset := fakeclient.NewSimpleClientset()
	dbClient := fakeoci.NewDatabaseClient()
	kubeclient := fakekube.NewSimpleClientset()

	adbAdapter := AutonomousDatabaseAdapter{}
	adbAdapter.clientset = clientset
	adbAdapter.kubeclient = kubeclient
	adbAdapter.dbClient = dbClient
	adbAdapter.seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

	comp, err := clientset.OciidentityV1alpha1().Compartments(fakeNs).Create(&compartment)
	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created compartment object %v", comp)

	newDb, err := adbAdapter.CreateObject(&adb)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created AutonomousDatabase object %v", newDb)

	if adbAdapter.IsExpectedType(adb) {
		t.Logf("Checked AutonomousDatabase type")
	}

	adbWithResource, err := adbAdapter.Create(&adb)

	if err != nil {
		t.Errorf("Got create AutonomousDatabase option error %v", err)
	}
	t.Logf("Created AutonomousDatabase resource %v", adbWithResource)

	adbWithResource, err = adbAdapter.Get(newDb)
	adbWithResource, err = adbAdapter.Update(newDb)
	adbWithResource, err = adbAdapter.Delete(newDb)
	eq := adbAdapter.Equivalent(newDb, newDb)
	if !eq {
		t.Errorf("should be equal")
	}
	_, err = adbAdapter.DependsOnRefs(newDb)
}

func NewFakeAutonomousDatabaseAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	dbClient := fakeoci.NewDatabaseClient()
	adbAdapter := AutonomousDatabaseAdapter{}
	adbAdapter.clientset = clientset
	adbAdapter.dbClient = dbClient
	return &adbAdapter
}
