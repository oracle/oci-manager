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
package versioned

import (
	cloudv1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/cloud.k8s.io/v1alpha1"
	ocicev1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ocice.oracle.com/v1alpha1"
	ocicorev1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ocicore.oracle.com/v1alpha1"
	ocidbv1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ocidb.oracle.com/v1alpha1"
	ociidentityv1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ociidentity.oracle.com/v1alpha1"
	ocilbv1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ocilb.oracle.com/v1alpha1"
	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	CloudV1alpha1() cloudv1alpha1.CloudV1alpha1Interface
	// Deprecated: please explicitly pick a version if possible.
	Cloud() cloudv1alpha1.CloudV1alpha1Interface
	OciceV1alpha1() ocicev1alpha1.OciceV1alpha1Interface
	// Deprecated: please explicitly pick a version if possible.
	Ocice() ocicev1alpha1.OciceV1alpha1Interface
	OcicoreV1alpha1() ocicorev1alpha1.OcicoreV1alpha1Interface
	// Deprecated: please explicitly pick a version if possible.
	Ocicore() ocicorev1alpha1.OcicoreV1alpha1Interface
	OcidbV1alpha1() ocidbv1alpha1.OcidbV1alpha1Interface
	// Deprecated: please explicitly pick a version if possible.
	Ocidb() ocidbv1alpha1.OcidbV1alpha1Interface
	OciidentityV1alpha1() ociidentityv1alpha1.OciidentityV1alpha1Interface
	// Deprecated: please explicitly pick a version if possible.
	Ociidentity() ociidentityv1alpha1.OciidentityV1alpha1Interface
	OcilbV1alpha1() ocilbv1alpha1.OcilbV1alpha1Interface
	// Deprecated: please explicitly pick a version if possible.
	Ocilb() ocilbv1alpha1.OcilbV1alpha1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	cloudV1alpha1       *cloudv1alpha1.CloudV1alpha1Client
	ociceV1alpha1       *ocicev1alpha1.OciceV1alpha1Client
	ocicoreV1alpha1     *ocicorev1alpha1.OcicoreV1alpha1Client
	ocidbV1alpha1       *ocidbv1alpha1.OcidbV1alpha1Client
	ociidentityV1alpha1 *ociidentityv1alpha1.OciidentityV1alpha1Client
	ocilbV1alpha1       *ocilbv1alpha1.OcilbV1alpha1Client
}

// CloudV1alpha1 retrieves the CloudV1alpha1Client
func (c *Clientset) CloudV1alpha1() cloudv1alpha1.CloudV1alpha1Interface {
	return c.cloudV1alpha1
}

// Deprecated: Cloud retrieves the default version of CloudClient.
// Please explicitly pick a version.
func (c *Clientset) Cloud() cloudv1alpha1.CloudV1alpha1Interface {
	return c.cloudV1alpha1
}

// OciceV1alpha1 retrieves the OciceV1alpha1Client
func (c *Clientset) OciceV1alpha1() ocicev1alpha1.OciceV1alpha1Interface {
	return c.ociceV1alpha1
}

// Deprecated: Ocice retrieves the default version of OciceClient.
// Please explicitly pick a version.
func (c *Clientset) Ocice() ocicev1alpha1.OciceV1alpha1Interface {
	return c.ociceV1alpha1
}

// OcicoreV1alpha1 retrieves the OcicoreV1alpha1Client
func (c *Clientset) OcicoreV1alpha1() ocicorev1alpha1.OcicoreV1alpha1Interface {
	return c.ocicoreV1alpha1
}

// Deprecated: Ocicore retrieves the default version of OcicoreClient.
// Please explicitly pick a version.
func (c *Clientset) Ocicore() ocicorev1alpha1.OcicoreV1alpha1Interface {
	return c.ocicoreV1alpha1
}

// OcidbV1alpha1 retrieves the OcidbV1alpha1Client
func (c *Clientset) OcidbV1alpha1() ocidbv1alpha1.OcidbV1alpha1Interface {
	return c.ocidbV1alpha1
}

// Deprecated: Ocidb retrieves the default version of OcidbClient.
// Please explicitly pick a version.
func (c *Clientset) Ocidb() ocidbv1alpha1.OcidbV1alpha1Interface {
	return c.ocidbV1alpha1
}

// OciidentityV1alpha1 retrieves the OciidentityV1alpha1Client
func (c *Clientset) OciidentityV1alpha1() ociidentityv1alpha1.OciidentityV1alpha1Interface {
	return c.ociidentityV1alpha1
}

// Deprecated: Ociidentity retrieves the default version of OciidentityClient.
// Please explicitly pick a version.
func (c *Clientset) Ociidentity() ociidentityv1alpha1.OciidentityV1alpha1Interface {
	return c.ociidentityV1alpha1
}

// OcilbV1alpha1 retrieves the OcilbV1alpha1Client
func (c *Clientset) OcilbV1alpha1() ocilbv1alpha1.OcilbV1alpha1Interface {
	return c.ocilbV1alpha1
}

// Deprecated: Ocilb retrieves the default version of OcilbClient.
// Please explicitly pick a version.
func (c *Clientset) Ocilb() ocilbv1alpha1.OcilbV1alpha1Interface {
	return c.ocilbV1alpha1
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.cloudV1alpha1, err = cloudv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.ociceV1alpha1, err = ocicev1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.ocicoreV1alpha1, err = ocicorev1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.ocidbV1alpha1, err = ocidbv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.ociidentityV1alpha1, err = ociidentityv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.ocilbV1alpha1, err = ocilbv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.cloudV1alpha1 = cloudv1alpha1.NewForConfigOrDie(c)
	cs.ociceV1alpha1 = ocicev1alpha1.NewForConfigOrDie(c)
	cs.ocicoreV1alpha1 = ocicorev1alpha1.NewForConfigOrDie(c)
	cs.ocidbV1alpha1 = ocidbv1alpha1.NewForConfigOrDie(c)
	cs.ociidentityV1alpha1 = ociidentityv1alpha1.NewForConfigOrDie(c)
	cs.ocilbV1alpha1 = ocilbv1alpha1.NewForConfigOrDie(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.cloudV1alpha1 = cloudv1alpha1.New(c)
	cs.ociceV1alpha1 = ocicev1alpha1.New(c)
	cs.ocicoreV1alpha1 = ocicorev1alpha1.New(c)
	cs.ocidbV1alpha1 = ocidbv1alpha1.New(c)
	cs.ociidentityV1alpha1 = ociidentityv1alpha1.New(c)
	cs.ocilbV1alpha1 = ocilbv1alpha1.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
