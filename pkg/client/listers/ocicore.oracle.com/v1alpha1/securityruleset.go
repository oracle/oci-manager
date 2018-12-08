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

// SecurityRuleSetLister helps list SecurityRuleSets.
type SecurityRuleSetLister interface {
	// List lists all SecurityRuleSets in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.SecurityRuleSet, err error)
	// SecurityRuleSets returns an object that can list and get SecurityRuleSets.
	SecurityRuleSets(namespace string) SecurityRuleSetNamespaceLister
	SecurityRuleSetListerExpansion
}

// securityRuleSetLister implements the SecurityRuleSetLister interface.
type securityRuleSetLister struct {
	indexer cache.Indexer
}

// NewSecurityRuleSetLister returns a new SecurityRuleSetLister.
func NewSecurityRuleSetLister(indexer cache.Indexer) SecurityRuleSetLister {
	return &securityRuleSetLister{indexer: indexer}
}

// List lists all SecurityRuleSets in the indexer.
func (s *securityRuleSetLister) List(selector labels.Selector) (ret []*v1alpha1.SecurityRuleSet, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.SecurityRuleSet))
	})
	return ret, err
}

// SecurityRuleSets returns an object that can list and get SecurityRuleSets.
func (s *securityRuleSetLister) SecurityRuleSets(namespace string) SecurityRuleSetNamespaceLister {
	return securityRuleSetNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// SecurityRuleSetNamespaceLister helps list and get SecurityRuleSets.
type SecurityRuleSetNamespaceLister interface {
	// List lists all SecurityRuleSets in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.SecurityRuleSet, err error)
	// Get retrieves the SecurityRuleSet from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.SecurityRuleSet, error)
	SecurityRuleSetNamespaceListerExpansion
}

// securityRuleSetNamespaceLister implements the SecurityRuleSetNamespaceLister
// interface.
type securityRuleSetNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all SecurityRuleSets in the indexer for a given namespace.
func (s securityRuleSetNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.SecurityRuleSet, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.SecurityRuleSet))
	})
	return ret, err
}

// Get retrieves the SecurityRuleSet from the indexer for a given namespace and name.
func (s securityRuleSetNamespaceLister) Get(name string) (*v1alpha1.SecurityRuleSet, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("securityruleset"), name)
	}
	return obj.(*v1alpha1.SecurityRuleSet), nil
}
