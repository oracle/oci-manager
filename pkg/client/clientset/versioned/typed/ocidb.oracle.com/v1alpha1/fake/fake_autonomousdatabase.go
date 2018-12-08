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
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocidb.oracle.com/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeAutonomousDatabases implements AutonomousDatabaseInterface
type FakeAutonomousDatabases struct {
	Fake *FakeOcidbV1alpha1
	ns   string
}

var autonomousdatabasesResource = schema.GroupVersionResource{Group: "ocidb.oracle.com", Version: "v1alpha1", Resource: "autonomousdatabases"}

var autonomousdatabasesKind = schema.GroupVersionKind{Group: "ocidb.oracle.com", Version: "v1alpha1", Kind: "AutonomousDatabase"}

// Get takes name of the autonomousDatabase, and returns the corresponding autonomousDatabase object, and an error if there is any.
func (c *FakeAutonomousDatabases) Get(name string, options v1.GetOptions) (result *v1alpha1.AutonomousDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(autonomousdatabasesResource, c.ns, name), &v1alpha1.AutonomousDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AutonomousDatabase), err
}

// List takes label and field selectors, and returns the list of AutonomousDatabases that match those selectors.
func (c *FakeAutonomousDatabases) List(opts v1.ListOptions) (result *v1alpha1.AutonomousDatabaseList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(autonomousdatabasesResource, autonomousdatabasesKind, c.ns, opts), &v1alpha1.AutonomousDatabaseList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.AutonomousDatabaseList{ListMeta: obj.(*v1alpha1.AutonomousDatabaseList).ListMeta}
	for _, item := range obj.(*v1alpha1.AutonomousDatabaseList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested autonomousDatabases.
func (c *FakeAutonomousDatabases) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(autonomousdatabasesResource, c.ns, opts))

}

// Create takes the representation of a autonomousDatabase and creates it.  Returns the server's representation of the autonomousDatabase, and an error, if there is any.
func (c *FakeAutonomousDatabases) Create(autonomousDatabase *v1alpha1.AutonomousDatabase) (result *v1alpha1.AutonomousDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(autonomousdatabasesResource, c.ns, autonomousDatabase), &v1alpha1.AutonomousDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AutonomousDatabase), err
}

// Update takes the representation of a autonomousDatabase and updates it. Returns the server's representation of the autonomousDatabase, and an error, if there is any.
func (c *FakeAutonomousDatabases) Update(autonomousDatabase *v1alpha1.AutonomousDatabase) (result *v1alpha1.AutonomousDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(autonomousdatabasesResource, c.ns, autonomousDatabase), &v1alpha1.AutonomousDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AutonomousDatabase), err
}

// Delete takes name of the autonomousDatabase and deletes it. Returns an error if one occurs.
func (c *FakeAutonomousDatabases) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(autonomousdatabasesResource, c.ns, name), &v1alpha1.AutonomousDatabase{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeAutonomousDatabases) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(autonomousdatabasesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.AutonomousDatabaseList{})
	return err
}

// Patch applies the patch and returns the patched autonomousDatabase.
func (c *FakeAutonomousDatabases) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.AutonomousDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(autonomousdatabasesResource, c.ns, name, data, subresources...), &v1alpha1.AutonomousDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AutonomousDatabase), err
}
