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
	v1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ociidentity.oracle.com/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeOciidentityV1alpha1 struct {
	*testing.Fake
}

func (c *FakeOciidentityV1alpha1) Compartments(namespace string) v1alpha1.CompartmentInterface {
	return &FakeCompartments{c, namespace}
}

func (c *FakeOciidentityV1alpha1) DynamicGroups(namespace string) v1alpha1.DynamicGroupInterface {
	return &FakeDynamicGroups{c, namespace}
}

func (c *FakeOciidentityV1alpha1) Policies(namespace string) v1alpha1.PolicyInterface {
	return &FakePolicies{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeOciidentityV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
