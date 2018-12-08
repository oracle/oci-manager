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

package resources

import (
	corev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	fakeclient "github.com/oracle/oci-manager/pkg/client/clientset/versioned/fake"
	informers "github.com/oracle/oci-manager/pkg/client/informers/externalversions"
	coreresources "github.com/oracle/oci-manager/pkg/controller/oci/resources/core"
	fakeoci "github.com/oracle/oci-manager/pkg/controller/oci/resources/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/util/workqueue"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"golang.org/x/time/rate"
)

const fakeNs = "fakeNs"

func TestControllerCreateHandle(t *testing.T) {

	vcntest1 := corev1alpha1.Vcn{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "vcn.test1",
			Namespace:  fakeNs,
			Finalizers: []string{"ocimanager"},
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       corev1alpha1.VirtualNetworkKind,
		},
		Spec: corev1alpha1.VcnSpec{
			DisplayName:    "testDisplay",
			CompartmentRef: "ocid1.compartment.oc1..aaaaaaaas3of2ieoysj25xjoqt6mw3ft5veqixo2fyoapobspc65novfpo6q",
		},
	}

	t.Log("Testing OCI controller CREATE with VCN Adapter")
	clientset := fakeclient.NewSimpleClientset()
	vcnClient := fakeoci.NewVcnClient()

	vcnAdapter := coreresources.NewVcnAdapterBasic(clientset, vcnClient)
	newVcn, err := vcnAdapter.CreateObject(&vcntest1)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Vcn object %v", newVcn)

	informerFactory := informers.NewSharedInformerFactory(clientset, 30*time.Second)
	stopCh := make(chan struct{})
	workQueues := make(map[string]workqueue.RateLimitingInterface)
	rateLimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(2*time.Second, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(float64(10)), 100)},
	)
	workQueues[vcnAdapter.Kind()] = workqueue.NewRateLimitingQueue(rateLimiter)

	t.Log("Starting controller informers")
	kubeclient := fake.NewSimpleClientset()
	controller := New(vcnAdapter, kubeclient, informerFactory, workQueues)
	controller.Run(stopCh)

	time.Sleep(1 * time.Second)

	realizedVcn, err := clientset.OcicoreV1alpha1().Vcns(fakeNs).Get("vcn.test1", metav1.GetOptions{})

	t.Logf("Vcn after update - %v", realizedVcn)

	if realizedVcn.GetResourceID() == "" {
		t.Errorf("Vcn resource id was not populated")
	} else {
		t.Logf("Vcn ocid populated - %s", realizedVcn.GetResourceID())
	}

	var stop struct{}
	stopCh <- stop

}

func TestControllerDeletHandle(t *testing.T) {

	timeNow := metav1.Now()
	vcntest1 := corev1alpha1.Vcn{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "vcn.test1",
			Namespace:         fakeNs,
			DeletionTimestamp: &timeNow,
			Finalizers:        []string{"ocimanager"},
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       corev1alpha1.VirtualNetworkKind,
		},
		Spec: corev1alpha1.VcnSpec{
			DisplayName:    "testDisplay",
			CompartmentRef: "ocid1.compartment.oc1..aaaaaaaas3of2ieoysj25xjoqt6mw3ft5veqixo2fyoapobspc65novfpo6q",
		},
	}

	t.Log("Testing OCI controller DELETE with VCN Adapter")
	clientset := fakeclient.NewSimpleClientset()
	vcnClient := fakeoci.NewVcnClient()

	vcnAdapter := coreresources.NewVcnAdapterBasic(clientset, vcnClient)
	newVcn, err := vcnAdapter.CreateObject(&vcntest1)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Vcn object %v", newVcn)

	informerFactory := informers.NewSharedInformerFactory(clientset, 30*time.Second)
	stopCh := make(chan struct{})
	workQueues := make(map[string]workqueue.RateLimitingInterface)

	rateLimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(2*time.Second, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(float64(10)), 100)},
	)

	workQueues[vcnAdapter.Kind()] = workqueue.NewRateLimitingQueue(rateLimiter)

	t.Log("Starting controller informers")
	kubeclient := fake.NewSimpleClientset()
	controller := New(vcnAdapter, kubeclient, informerFactory, workQueues)
	controller.Run(stopCh)

	time.Sleep(1 * time.Second)

	realizedVcn, err := clientset.OcicoreV1alpha1().Vcns(fakeNs).Get("vcn.test1", metav1.GetOptions{})

	t.Logf("Vcn after delete - %v", realizedVcn)
	if len(realizedVcn.ObjectMeta.Finalizers) > 0 {
		t.Errorf("Finalizers should be removed")
	} else {
		t.Logf("Finalizers has been removed by controller")
	}

	var stop struct{}
	stopCh <- stop

}
