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
	// Backends returns a BackendInformer.
	Backends() BackendInformer
	// BackendSets returns a BackendSetInformer.
	BackendSets() BackendSetInformer
	// Certificates returns a CertificateInformer.
	Certificates() CertificateInformer
	// Listeners returns a ListenerInformer.
	Listeners() ListenerInformer
	// LoadBalancers returns a LoadBalancerInformer.
	LoadBalancers() LoadBalancerInformer
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

// Backends returns a BackendInformer.
func (v *version) Backends() BackendInformer {
	return &backendInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// BackendSets returns a BackendSetInformer.
func (v *version) BackendSets() BackendSetInformer {
	return &backendSetInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Certificates returns a CertificateInformer.
func (v *version) Certificates() CertificateInformer {
	return &certificateInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Listeners returns a ListenerInformer.
func (v *version) Listeners() ListenerInformer {
	return &listenerInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// LoadBalancers returns a LoadBalancerInformer.
func (v *version) LoadBalancers() LoadBalancerInformer {
	return &loadBalancerInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
