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
	internalinterfaces "github.com/oracle/oci-manager/pkg/client/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// Clusters returns a ClusterInformer.
	Clusters() ClusterInformer
	// Computes returns a ComputeInformer.
	Computes() ComputeInformer
	// Cpods returns a CpodInformer.
	Cpods() CpodInformer
	// LoadBalancers returns a LoadBalancerInformer.
	LoadBalancers() LoadBalancerInformer
	// Networks returns a NetworkInformer.
	Networks() NetworkInformer
	// Securities returns a SecurityInformer.
	Securities() SecurityInformer
	// Storages returns a StorageInformer.
	Storages() StorageInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// Clusters returns a ClusterInformer.
func (v *version) Clusters() ClusterInformer {
	return &clusterInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Computes returns a ComputeInformer.
func (v *version) Computes() ComputeInformer {
	return &computeInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Cpods returns a CpodInformer.
func (v *version) Cpods() CpodInformer {
	return &cpodInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// LoadBalancers returns a LoadBalancerInformer.
func (v *version) LoadBalancers() LoadBalancerInformer {
	return &loadBalancerInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Networks returns a NetworkInformer.
func (v *version) Networks() NetworkInformer {
	return &networkInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Securities returns a SecurityInformer.
func (v *version) Securities() SecurityInformer {
	return &securityInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Storages returns a StorageInformer.
func (v *version) Storages() StorageInformer {
	return &storageInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
