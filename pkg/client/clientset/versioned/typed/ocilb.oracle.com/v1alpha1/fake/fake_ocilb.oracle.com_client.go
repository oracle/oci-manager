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
package fake

import (
	v1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ocilb.oracle.com/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeOcilbV1alpha1 struct {
	*testing.Fake
}

func (c *FakeOcilbV1alpha1) Backends(namespace string) v1alpha1.BackendInterface {
	return &FakeBackends{c, namespace}
}

func (c *FakeOcilbV1alpha1) BackendSets(namespace string) v1alpha1.BackendSetInterface {
	return &FakeBackendSets{c, namespace}
}

func (c *FakeOcilbV1alpha1) Certificates(namespace string) v1alpha1.CertificateInterface {
	return &FakeCertificates{c, namespace}
}

func (c *FakeOcilbV1alpha1) Listeners(namespace string) v1alpha1.ListenerInterface {
	return &FakeListeners{c, namespace}
}

func (c *FakeOcilbV1alpha1) LoadBalancers(namespace string) v1alpha1.LoadBalancerInterface {
	return &FakeLoadBalancers{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeOcilbV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
