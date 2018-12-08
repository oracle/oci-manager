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

// DhcpOptionLister helps list DhcpOptions.
type DhcpOptionLister interface {
	// List lists all DhcpOptions in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.DhcpOption, err error)
	// DhcpOptions returns an object that can list and get DhcpOptions.
	DhcpOptions(namespace string) DhcpOptionNamespaceLister
	DhcpOptionListerExpansion
}

// dhcpOptionLister implements the DhcpOptionLister interface.
type dhcpOptionLister struct {
	indexer cache.Indexer
}

// NewDhcpOptionLister returns a new DhcpOptionLister.
func NewDhcpOptionLister(indexer cache.Indexer) DhcpOptionLister {
	return &dhcpOptionLister{indexer: indexer}
}

// List lists all DhcpOptions in the indexer.
func (s *dhcpOptionLister) List(selector labels.Selector) (ret []*v1alpha1.DhcpOption, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.DhcpOption))
	})
	return ret, err
}

// DhcpOptions returns an object that can list and get DhcpOptions.
func (s *dhcpOptionLister) DhcpOptions(namespace string) DhcpOptionNamespaceLister {
	return dhcpOptionNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// DhcpOptionNamespaceLister helps list and get DhcpOptions.
type DhcpOptionNamespaceLister interface {
	// List lists all DhcpOptions in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.DhcpOption, err error)
	// Get retrieves the DhcpOption from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.DhcpOption, error)
	DhcpOptionNamespaceListerExpansion
}

// dhcpOptionNamespaceLister implements the DhcpOptionNamespaceLister
// interface.
type dhcpOptionNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all DhcpOptions in the indexer for a given namespace.
func (s dhcpOptionNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.DhcpOption, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.DhcpOption))
	})
	return ret, err
}

// Get retrieves the DhcpOption from the indexer for a given namespace and name.
func (s dhcpOptionNamespaceLister) Get(name string) (*v1alpha1.DhcpOption, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("dhcpoption"), name)
	}
	return obj.(*v1alpha1.DhcpOption), nil
}
