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
package v1alpha1

import (
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned/scheme"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
)

type OcicoreV1alpha1Interface interface {
	RESTClient() rest.Interface
	DhcpOptionsGetter
	InstancesGetter
	InternetGatewaiesGetter
	RouteTablesGetter
	SecurityRuleSetsGetter
	SubnetsGetter
	VcnsGetter
	VolumesGetter
	VolumeBackupsGetter
}

// OcicoreV1alpha1Client is used to interact with features provided by the ocicore.oracle.com group.
type OcicoreV1alpha1Client struct {
	restClient rest.Interface
}

func (c *OcicoreV1alpha1Client) DhcpOptions(namespace string) DhcpOptionInterface {
	return newDhcpOptions(c, namespace)
}

func (c *OcicoreV1alpha1Client) Instances(namespace string) InstanceInterface {
	return newInstances(c, namespace)
}

func (c *OcicoreV1alpha1Client) InternetGatewaies(namespace string) InternetGatewayInterface {
	return newInternetGatewaies(c, namespace)
}

func (c *OcicoreV1alpha1Client) RouteTables(namespace string) RouteTableInterface {
	return newRouteTables(c, namespace)
}

func (c *OcicoreV1alpha1Client) SecurityRuleSets(namespace string) SecurityRuleSetInterface {
	return newSecurityRuleSets(c, namespace)
}

func (c *OcicoreV1alpha1Client) Subnets(namespace string) SubnetInterface {
	return newSubnets(c, namespace)
}

func (c *OcicoreV1alpha1Client) Vcns(namespace string) VcnInterface {
	return newVcns(c, namespace)
}

func (c *OcicoreV1alpha1Client) Volumes(namespace string) VolumeInterface {
	return newVolumes(c, namespace)
}

func (c *OcicoreV1alpha1Client) VolumeBackups(namespace string) VolumeBackupInterface {
	return newVolumeBackups(c, namespace)
}

// NewForConfig creates a new OcicoreV1alpha1Client for the given config.
func NewForConfig(c *rest.Config) (*OcicoreV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &OcicoreV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new OcicoreV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *OcicoreV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new OcicoreV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *OcicoreV1alpha1Client {
	return &OcicoreV1alpha1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *OcicoreV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
