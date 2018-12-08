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
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocidb.oracle.com/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// AutonomousDatabaseLister helps list AutonomousDatabases.
type AutonomousDatabaseLister interface {
	// List lists all AutonomousDatabases in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.AutonomousDatabase, err error)
	// AutonomousDatabases returns an object that can list and get AutonomousDatabases.
	AutonomousDatabases(namespace string) AutonomousDatabaseNamespaceLister
	AutonomousDatabaseListerExpansion
}

// autonomousDatabaseLister implements the AutonomousDatabaseLister interface.
type autonomousDatabaseLister struct {
	indexer cache.Indexer
}

// NewAutonomousDatabaseLister returns a new AutonomousDatabaseLister.
func NewAutonomousDatabaseLister(indexer cache.Indexer) AutonomousDatabaseLister {
	return &autonomousDatabaseLister{indexer: indexer}
}

// List lists all AutonomousDatabases in the indexer.
func (s *autonomousDatabaseLister) List(selector labels.Selector) (ret []*v1alpha1.AutonomousDatabase, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.AutonomousDatabase))
	})
	return ret, err
}

// AutonomousDatabases returns an object that can list and get AutonomousDatabases.
func (s *autonomousDatabaseLister) AutonomousDatabases(namespace string) AutonomousDatabaseNamespaceLister {
	return autonomousDatabaseNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// AutonomousDatabaseNamespaceLister helps list and get AutonomousDatabases.
type AutonomousDatabaseNamespaceLister interface {
	// List lists all AutonomousDatabases in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.AutonomousDatabase, err error)
	// Get retrieves the AutonomousDatabase from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.AutonomousDatabase, error)
	AutonomousDatabaseNamespaceListerExpansion
}

// autonomousDatabaseNamespaceLister implements the AutonomousDatabaseNamespaceLister
// interface.
type autonomousDatabaseNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all AutonomousDatabases in the indexer for a given namespace.
func (s autonomousDatabaseNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.AutonomousDatabase, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.AutonomousDatabase))
	})
	return ret, err
}

// Get retrieves the AutonomousDatabase from the indexer for a given namespace and name.
func (s autonomousDatabaseNamespaceLister) Get(name string) (*v1alpha1.AutonomousDatabase, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("autonomousdatabase"), name)
	}
	return obj.(*v1alpha1.AutonomousDatabase), nil
}
