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
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/cloud.k8s.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ComputeLister helps list Computes.
type ComputeLister interface {
	// List lists all Computes in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.Compute, err error)
	// Computes returns an object that can list and get Computes.
	Computes(namespace string) ComputeNamespaceLister
	ComputeListerExpansion
}

// computeLister implements the ComputeLister interface.
type computeLister struct {
	indexer cache.Indexer
}

// NewComputeLister returns a new ComputeLister.
func NewComputeLister(indexer cache.Indexer) ComputeLister {
	return &computeLister{indexer: indexer}
}

// List lists all Computes in the indexer.
func (s *computeLister) List(selector labels.Selector) (ret []*v1alpha1.Compute, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Compute))
	})
	return ret, err
}

// Computes returns an object that can list and get Computes.
func (s *computeLister) Computes(namespace string) ComputeNamespaceLister {
	return computeNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ComputeNamespaceLister helps list and get Computes.
type ComputeNamespaceLister interface {
	// List lists all Computes in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.Compute, err error)
	// Get retrieves the Compute from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.Compute, error)
	ComputeNamespaceListerExpansion
}

// computeNamespaceLister implements the ComputeNamespaceLister
// interface.
type computeNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Computes in the indexer for a given namespace.
func (s computeNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Compute, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Compute))
	})
	return ret, err
}

// Get retrieves the Compute from the indexer for a given namespace and name.
func (s computeNamespaceLister) Get(name string) (*v1alpha1.Compute, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("compute"), name)
	}
	return obj.(*v1alpha1.Compute), nil
}
