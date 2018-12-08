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
package externalversions

import (
	"fmt"
	v1alpha1 "github.com/oracle/oci-manager/pkg/apis/cloud.k8s.io/v1alpha1"
	ocice_oracle_com_v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocice.oracle.com/v1alpha1"
	ocicore_oracle_com_v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	ocidb_oracle_com_v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocidb.oracle.com/v1alpha1"
	ociidentity_oracle_com_v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com/v1alpha1"
	ocilb_oracle_com_v1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocilb.oracle.com/v1alpha1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	cache "k8s.io/client-go/tools/cache"
)

// GenericInformer is type of SharedIndexInformer which will locate and delegate to other
// sharedInformers based on type
type GenericInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}

type genericInformer struct {
	informer cache.SharedIndexInformer
	resource schema.GroupResource
}

// Informer returns the SharedIndexInformer.
func (f *genericInformer) Informer() cache.SharedIndexInformer {
	return f.informer
}

// Lister returns the GenericLister.
func (f *genericInformer) Lister() cache.GenericLister {
	return cache.NewGenericLister(f.Informer().GetIndexer(), f.resource)
}

// ForResource gives generic access to a shared informer of the matching type
// TODO extend this to unknown resources with a client pool
func (f *sharedInformerFactory) ForResource(resource schema.GroupVersionResource) (GenericInformer, error) {
	switch resource {
	// Group=cloud.k8s.io, Version=v1alpha1
	case v1alpha1.SchemeGroupVersion.WithResource("clusters"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cloud().V1alpha1().Clusters().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("computes"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cloud().V1alpha1().Computes().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("cpods"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cloud().V1alpha1().Cpods().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("loadbalancers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cloud().V1alpha1().LoadBalancers().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("networks"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cloud().V1alpha1().Networks().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("securities"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cloud().V1alpha1().Securities().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("storages"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cloud().V1alpha1().Storages().Informer()}, nil

		// Group=ocice.oracle.com, Version=v1alpha1
	case ocice_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("clusters"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ocice().V1alpha1().Clusters().Informer()}, nil
	case ocice_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("nodepools"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ocice().V1alpha1().NodePools().Informer()}, nil

		// Group=ocicore.oracle.com, Version=v1alpha1
	case ocicore_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("dhcpoptions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ocicore().V1alpha1().DhcpOptions().Informer()}, nil
	case ocicore_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("instances"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ocicore().V1alpha1().Instances().Informer()}, nil
	case ocicore_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("internetgatewaies"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ocicore().V1alpha1().InternetGatewaies().Informer()}, nil
	case ocicore_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("routetables"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ocicore().V1alpha1().RouteTables().Informer()}, nil
	case ocicore_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("securityrulesets"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ocicore().V1alpha1().SecurityRuleSets().Informer()}, nil
	case ocicore_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("subnets"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ocicore().V1alpha1().Subnets().Informer()}, nil
	case ocicore_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("vcns"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ocicore().V1alpha1().Vcns().Informer()}, nil
	case ocicore_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("volumes"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ocicore().V1alpha1().Volumes().Informer()}, nil
	case ocicore_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("volumebackups"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ocicore().V1alpha1().VolumeBackups().Informer()}, nil

		// Group=ocidb.oracle.com, Version=v1alpha1
	case ocidb_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("autonomousdatabases"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ocidb().V1alpha1().AutonomousDatabases().Informer()}, nil

		// Group=ociidentity.oracle.com, Version=v1alpha1
	case ociidentity_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("compartments"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ociidentity().V1alpha1().Compartments().Informer()}, nil
	case ociidentity_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("dynamicgroups"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ociidentity().V1alpha1().DynamicGroups().Informer()}, nil
	case ociidentity_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("policies"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ociidentity().V1alpha1().Policies().Informer()}, nil

		// Group=ocilb.oracle.com, Version=v1alpha1
	case ocilb_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("backends"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ocilb().V1alpha1().Backends().Informer()}, nil
	case ocilb_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("backendsets"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ocilb().V1alpha1().BackendSets().Informer()}, nil
	case ocilb_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("certificates"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ocilb().V1alpha1().Certificates().Informer()}, nil
	case ocilb_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("listeners"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ocilb().V1alpha1().Listeners().Informer()}, nil
	case ocilb_oracle_com_v1alpha1.SchemeGroupVersion.WithResource("loadbalancers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ocilb().V1alpha1().LoadBalancers().Informer()}, nil

	}

	return nil, fmt.Errorf("no informer found for %v", resource)
}
