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
	v1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/cloud.k8s.io/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeCloudV1alpha1 struct {
	*testing.Fake
}

func (c *FakeCloudV1alpha1) Clusters(namespace string) v1alpha1.ClusterInterface {
	return &FakeClusters{c, namespace}
}

func (c *FakeCloudV1alpha1) Computes(namespace string) v1alpha1.ComputeInterface {
	return &FakeComputes{c, namespace}
}

func (c *FakeCloudV1alpha1) Cpods(namespace string) v1alpha1.CpodInterface {
	return &FakeCpods{c, namespace}
}

func (c *FakeCloudV1alpha1) LoadBalancers(namespace string) v1alpha1.LoadBalancerInterface {
	return &FakeLoadBalancers{c, namespace}
}

func (c *FakeCloudV1alpha1) Networks(namespace string) v1alpha1.NetworkInterface {
	return &FakeNetworks{c, namespace}
}

func (c *FakeCloudV1alpha1) Securities(namespace string) v1alpha1.SecurityInterface {
	return &FakeSecurities{c, namespace}
}

func (c *FakeCloudV1alpha1) Storages(namespace string) v1alpha1.StorageInterface {
	return &FakeStorages{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeCloudV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
