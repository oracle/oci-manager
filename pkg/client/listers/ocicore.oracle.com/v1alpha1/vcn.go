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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// VcnLister helps list Vcns.
type VcnLister interface {
	// List lists all Vcns in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.Vcn, err error)
	// Vcns returns an object that can list and get Vcns.
	Vcns(namespace string) VcnNamespaceLister
	VcnListerExpansion
}

// vcnLister implements the VcnLister interface.
type vcnLister struct {
	indexer cache.Indexer
}

// NewVcnLister returns a new VcnLister.
func NewVcnLister(indexer cache.Indexer) VcnLister {
	return &vcnLister{indexer: indexer}
}

// List lists all Vcns in the indexer.
func (s *vcnLister) List(selector labels.Selector) (ret []*v1alpha1.Vcn, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Vcn))
	})
	return ret, err
}

// Vcns returns an object that can list and get Vcns.
func (s *vcnLister) Vcns(namespace string) VcnNamespaceLister {
	return vcnNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// VcnNamespaceLister helps list and get Vcns.
type VcnNamespaceLister interface {
	// List lists all Vcns in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.Vcn, err error)
	// Get retrieves the Vcn from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.Vcn, error)
	VcnNamespaceListerExpansion
}

// vcnNamespaceLister implements the VcnNamespaceLister
// interface.
type vcnNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Vcns in the indexer for a given namespace.
func (s vcnNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Vcn, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Vcn))
	})
	return ret, err
}

// Get retrieves the Vcn from the indexer for a given namespace and name.
func (s vcnNamespaceLister) Get(name string) (*v1alpha1.Vcn, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("vcn"), name)
	}
	return obj.(*v1alpha1.Vcn), nil
}
