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
	certificatetest1 = lbv1alpha1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "certificate.test1",
			Namespace: fakeNs,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ocicore.oracle.com/v1alpha1",
			Kind:       lbv1alpha1.CertificateKind,
		},
		Spec: lbv1alpha1.CertificateSpec{
			PublicCertificate: "bla",
			PrivateKey:        "bla",
			Passphrase:        "bla",
		},
	}
)

func TestCertificateResourceBasic(t *testing.T) {

	t.Log("Testing certificate_resource")
	clientset := fakeclient.NewSimpleClientset()
	lbClient := fakeoci.NewLoadBalancerClient()

	certificateAdapter := CertificateAdapter{}
	certificateAdapter.clientset = clientset
	certificateAdapter.lbClient = lbClient

	lb, err := clientset.OcilbV1alpha1().LoadBalancers(fakeNs).Create(&lbTest1)
	if err != nil {
		t.Errorf("Got lb error %v", err)
	}
	t.Logf("Created lb object %v", lb)

	newCertificate, err := certificateAdapter.CreateObject(&certificatetest1)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Certificate object %v", newCertificate)

	if certificateAdapter.IsExpectedType(newCertificate) {
		t.Logf("Checked Certificate type")
	}

	certificateWithResource, err := certificateAdapter.Create(newCertificate)

	if err != nil {
		t.Errorf("Got error %v", err)
	}
	t.Logf("Created Certificate resource %v", certificateWithResource)

	certificateWithResource, err = certificateAdapter.Get(certificateWithResource)
	_, err = certificateAdapter.Update(certificateWithResource)
	_, err = certificateAdapter.Delete(certificateWithResource)
	_, err = certificateAdapter.DependsOnRefs(certificateWithResource)

}

func NewFakeCertificateAdapter(clientset versioned.Interface) resourcescommon.ResourceTypeAdapter {
	certificateClient := fakeoci.NewLoadBalancerClient()
	certificateAdapter := CertificateAdapter{}
	certificateAdapter.clientset = clientset
	certificateAdapter.lbClient = certificateClient
	return &certificateAdapter
}
