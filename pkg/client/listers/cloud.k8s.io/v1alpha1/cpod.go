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

// CpodLister helps list Cpods.
type CpodLister interface {
	// List lists all Cpods in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.Cpod, err error)
	// Cpods returns an object that can list and get Cpods.
	Cpods(namespace string) CpodNamespaceLister
	CpodListerExpansion
}

// cpodLister implements the CpodLister interface.
type cpodLister struct {
	indexer cache.Indexer
}

// NewCpodLister returns a new CpodLister.
func NewCpodLister(indexer cache.Indexer) CpodLister {
	return &cpodLister{indexer: indexer}
}

// List lists all Cpods in the indexer.
func (s *cpodLister) List(selector labels.Selector) (ret []*v1alpha1.Cpod, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Cpod))
	})
	return ret, err
}

// Cpods returns an object that can list and get Cpods.
func (s *cpodLister) Cpods(namespace string) CpodNamespaceLister {
	return cpodNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// CpodNamespaceLister helps list and get Cpods.
type CpodNamespaceLister interface {
	// List lists all Cpods in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.Cpod, err error)
	// Get retrieves the Cpod from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.Cpod, error)
	CpodNamespaceListerExpansion
}

// cpodNamespaceLister implements the CpodNamespaceLister
// interface.
type cpodNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Cpods in the indexer for a given namespace.
func (s cpodNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Cpod, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Cpod))
	})
	return ret, err
}

// Get retrieves the Cpod from the indexer for a given namespace and name.
func (s cpodNamespaceLister) Get(name string) (*v1alpha1.Cpod, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("cpod"), name)
	}
	return obj.(*v1alpha1.Cpod), nil
}
