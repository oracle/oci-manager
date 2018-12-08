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
	clientset "github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	cloudv1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/cloud.k8s.io/v1alpha1"
	fakecloudv1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/cloud.k8s.io/v1alpha1/fake"
	ocicev1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ocice.oracle.com/v1alpha1"
	fakeocicev1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ocice.oracle.com/v1alpha1/fake"
	ocicorev1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ocicore.oracle.com/v1alpha1"
	fakeocicorev1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ocicore.oracle.com/v1alpha1/fake"
	ocidbv1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ocidb.oracle.com/v1alpha1"
	fakeocidbv1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ocidb.oracle.com/v1alpha1/fake"
	ociidentityv1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ociidentity.oracle.com/v1alpha1"
	fakeociidentityv1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ociidentity.oracle.com/v1alpha1/fake"
	ocilbv1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ocilb.oracle.com/v1alpha1"
	fakeocilbv1alpha1 "github.com/oracle/oci-manager/pkg/client/clientset/versioned/typed/ocilb.oracle.com/v1alpha1/fake"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/testing"
)

// NewSimpleClientset returns a clientset that will respond with the provided objects.
// It's backed by a very simple object tracker that processes creates, updates and deletions as-is,
// without applying any validations and/or defaults. It shouldn't be considered a replacement
// for a real clientset and is mostly useful in simple unit tests.
func NewSimpleClientset(objects ...runtime.Object) *Clientset {
	o := testing.NewObjectTracker(scheme, codecs.UniversalDecoder())
	for _, obj := range objects {
		if err := o.Add(obj); err != nil {
			panic(err)
		}
	}

	cs := &Clientset{}
	cs.discovery = &fakediscovery.FakeDiscovery{Fake: &cs.Fake}
	cs.AddReactor("*", "*", testing.ObjectReaction(o))
	cs.AddWatchReactor("*", func(action testing.Action) (handled bool, ret watch.Interface, err error) {
		gvr := action.GetResource()
		ns := action.GetNamespace()
		watch, err := o.Watch(gvr, ns)
		if err != nil {
			return false, nil, err
		}
		return true, watch, nil
	})

	return cs
}

// Clientset implements clientset.Interface. Meant to be embedded into a
// struct to get a default implementation. This makes faking out just the method
// you want to test easier.
type Clientset struct {
	testing.Fake
	discovery *fakediscovery.FakeDiscovery
}

func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	return c.discovery
}

var _ clientset.Interface = &Clientset{}

// CloudV1alpha1 retrieves the CloudV1alpha1Client
func (c *Clientset) CloudV1alpha1() cloudv1alpha1.CloudV1alpha1Interface {
	return &fakecloudv1alpha1.FakeCloudV1alpha1{Fake: &c.Fake}
}

// Cloud retrieves the CloudV1alpha1Client
func (c *Clientset) Cloud() cloudv1alpha1.CloudV1alpha1Interface {
	return &fakecloudv1alpha1.FakeCloudV1alpha1{Fake: &c.Fake}
}

// OciceV1alpha1 retrieves the OciceV1alpha1Client
func (c *Clientset) OciceV1alpha1() ocicev1alpha1.OciceV1alpha1Interface {
	return &fakeocicev1alpha1.FakeOciceV1alpha1{Fake: &c.Fake}
}

// Ocice retrieves the OciceV1alpha1Client
func (c *Clientset) Ocice() ocicev1alpha1.OciceV1alpha1Interface {
	return &fakeocicev1alpha1.FakeOciceV1alpha1{Fake: &c.Fake}
}

// OcicoreV1alpha1 retrieves the OcicoreV1alpha1Client
func (c *Clientset) OcicoreV1alpha1() ocicorev1alpha1.OcicoreV1alpha1Interface {
	return &fakeocicorev1alpha1.FakeOcicoreV1alpha1{Fake: &c.Fake}
}

// Ocicore retrieves the OcicoreV1alpha1Client
func (c *Clientset) Ocicore() ocicorev1alpha1.OcicoreV1alpha1Interface {
	return &fakeocicorev1alpha1.FakeOcicoreV1alpha1{Fake: &c.Fake}
}

// OcidbV1alpha1 retrieves the OcidbV1alpha1Client
func (c *Clientset) OcidbV1alpha1() ocidbv1alpha1.OcidbV1alpha1Interface {
	return &fakeocidbv1alpha1.FakeOcidbV1alpha1{Fake: &c.Fake}
}

// Ocidb retrieves the OcidbV1alpha1Client
func (c *Clientset) Ocidb() ocidbv1alpha1.OcidbV1alpha1Interface {
	return &fakeocidbv1alpha1.FakeOcidbV1alpha1{Fake: &c.Fake}
}

// OciidentityV1alpha1 retrieves the OciidentityV1alpha1Client
func (c *Clientset) OciidentityV1alpha1() ociidentityv1alpha1.OciidentityV1alpha1Interface {
	return &fakeociidentityv1alpha1.FakeOciidentityV1alpha1{Fake: &c.Fake}
}

// Ociidentity retrieves the OciidentityV1alpha1Client
func (c *Clientset) Ociidentity() ociidentityv1alpha1.OciidentityV1alpha1Interface {
	return &fakeociidentityv1alpha1.FakeOciidentityV1alpha1{Fake: &c.Fake}
}

// OcilbV1alpha1 retrieves the OcilbV1alpha1Client
func (c *Clientset) OcilbV1alpha1() ocilbv1alpha1.OcilbV1alpha1Interface {
	return &fakeocilbv1alpha1.FakeOcilbV1alpha1{Fake: &c.Fake}
}

// Ocilb retrieves the OcilbV1alpha1Client
func (c *Clientset) Ocilb() ocilbv1alpha1.OcilbV1alpha1Interface {
	return &fakeocilbv1alpha1.FakeOcilbV1alpha1{Fake: &c.Fake}
}
