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
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocilb.oracle.com/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// BackendSetLister helps list BackendSets.
type BackendSetLister interface {
	// List lists all BackendSets in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.BackendSet, err error)
	// BackendSets returns an object that can list and get BackendSets.
	BackendSets(namespace string) BackendSetNamespaceLister
	BackendSetListerExpansion
}

// backendSetLister implements the BackendSetLister interface.
type backendSetLister struct {
	indexer cache.Indexer
}

// NewBackendSetLister returns a new BackendSetLister.
func NewBackendSetLister(indexer cache.Indexer) BackendSetLister {
	return &backendSetLister{indexer: indexer}
}

// List lists all BackendSets in the indexer.
func (s *backendSetLister) List(selector labels.Selector) (ret []*v1alpha1.BackendSet, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.BackendSet))
	})
	return ret, err
}

// BackendSets returns an object that can list and get BackendSets.
func (s *backendSetLister) BackendSets(namespace string) BackendSetNamespaceLister {
	return backendSetNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// BackendSetNamespaceLister helps list and get BackendSets.
type BackendSetNamespaceLister interface {
	// List lists all BackendSets in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.BackendSet, err error)
	// Get retrieves the BackendSet from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.BackendSet, error)
	BackendSetNamespaceListerExpansion
}

// backendSetNamespaceLister implements the BackendSetNamespaceLister
// interface.
type backendSetNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all BackendSets in the indexer for a given namespace.
func (s backendSetNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.BackendSet, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.BackendSet))
	})
	return ret, err
}

// Get retrieves the BackendSet from the indexer for a given namespace and name.
func (s backendSetNamespaceLister) Get(name string) (*v1alpha1.BackendSet, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("backendset"), name)
	}
	return obj.(*v1alpha1.BackendSet), nil
}
