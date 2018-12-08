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
	v1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ocicore.oracle.com/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeOcicoreV1alpha1 struct {
	*testing.Fake
}

func (c *FakeOcicoreV1alpha1) DhcpOptions(namespace string) v1alpha1.DhcpOptionInterface {
	return &FakeDhcpOptions{c, namespace}
}

func (c *FakeOcicoreV1alpha1) Instances(namespace string) v1alpha1.InstanceInterface {
	return &FakeInstances{c, namespace}
}

func (c *FakeOcicoreV1alpha1) InternetGatewaies(namespace string) v1alpha1.InternetGatewayInterface {
	return &FakeInternetGatewaies{c, namespace}
}

func (c *FakeOcicoreV1alpha1) RouteTables(namespace string) v1alpha1.RouteTableInterface {
	return &FakeRouteTables{c, namespace}
}

func (c *FakeOcicoreV1alpha1) SecurityRuleSets(namespace string) v1alpha1.SecurityRuleSetInterface {
	return &FakeSecurityRuleSets{c, namespace}
}

func (c *FakeOcicoreV1alpha1) Subnets(namespace string) v1alpha1.SubnetInterface {
	return &FakeSubnets{c, namespace}
}

func (c *FakeOcicoreV1alpha1) Vcns(namespace string) v1alpha1.VcnInterface {
	return &FakeVcns{c, namespace}
}

func (c *FakeOcicoreV1alpha1) Volumes(namespace string) v1alpha1.VolumeInterface {
	return &FakeVolumes{c, namespace}
}

func (c *FakeOcicoreV1alpha1) VolumeBackups(namespace string) v1alpha1.VolumeBackupInterface {
	return &FakeVolumeBackups{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeOcicoreV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
