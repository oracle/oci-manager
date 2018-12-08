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

// SecurityLister helps list Securities.
type SecurityLister interface {
	// List lists all Securities in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.Security, err error)
	// Securities returns an object that can list and get Securities.
	Securities(namespace string) SecurityNamespaceLister
	SecurityListerExpansion
}

// securityLister implements the SecurityLister interface.
type securityLister struct {
	indexer cache.Indexer
}

// NewSecurityLister returns a new SecurityLister.
func NewSecurityLister(indexer cache.Indexer) SecurityLister {
	return &securityLister{indexer: indexer}
}

// List lists all Securities in the indexer.
func (s *securityLister) List(selector labels.Selector) (ret []*v1alpha1.Security, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Security))
	})
	return ret, err
}

// Securities returns an object that can list and get Securities.
func (s *securityLister) Securities(namespace string) SecurityNamespaceLister {
	return securityNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// SecurityNamespaceLister helps list and get Securities.
type SecurityNamespaceLister interface {
	// List lists all Securities in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.Security, err error)
	// Get retrieves the Security from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.Security, error)
	SecurityNamespaceListerExpansion
}

// securityNamespaceLister implements the SecurityNamespaceLister
// interface.
type securityNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Securities in the indexer for a given namespace.
func (s securityNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Security, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Security))
	})
	return ret, err
}

// Get retrieves the Security from the indexer for a given namespace and name.
func (s securityNamespaceLister) Get(name string) (*v1alpha1.Security, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("security"), name)
	}
	return obj.(*v1alpha1.Security), nil
}
