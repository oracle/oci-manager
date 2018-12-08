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

package common

import (
	"fmt"
	ocicev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocice.oracle.com/v1alpha1"
	ocicorev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocicore.oracle.com/v1alpha1"
	ociidentityv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ociidentity.oracle.com/v1alpha1"
	ocilbv1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocilb.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// UpdateForResource is a generic common method to update the object with the corresponding resource
func UpdateForResource(clientset versioned.Interface, resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	switch resource {
	case ociidentityv1alpha1.SchemeGroupVersion.WithResource("compartments"):
		object := obj.(*ociidentityv1alpha1.Compartment)
		return clientset.OciidentityV1alpha1().Compartments(object.Namespace).Update(object)
	case ociidentityv1alpha1.SchemeGroupVersion.WithResource("policies"):
		object := obj.(*ociidentityv1alpha1.Policy)
		return clientset.OciidentityV1alpha1().Policies(object.Namespace).Update(object)
	case ocicorev1alpha1.SchemeGroupVersion.WithResource("instances"):
		object := obj.(*ocicorev1alpha1.Instance)
		return clientset.OcicoreV1alpha1().Instances(object.Namespace).Update(object)
	case ocicorev1alpha1.SchemeGroupVersion.WithResource("internetgatewaies"):
		object := obj.(*ocicorev1alpha1.InternetGateway)
		return clientset.OcicoreV1alpha1().InternetGatewaies(object.Namespace).Update(object)
	case ocicorev1alpha1.SchemeGroupVersion.WithResource("routetables"):
		object := obj.(*ocicorev1alpha1.RouteTable)
		return clientset.OcicoreV1alpha1().RouteTables(object.Namespace).Update(object)
	case ocicorev1alpha1.SchemeGroupVersion.WithResource("securityrulesets"):
		object := obj.(*ocicorev1alpha1.SecurityRuleSet)
		return clientset.OcicoreV1alpha1().SecurityRuleSets(object.Namespace).Update(object)
	case ocicorev1alpha1.SchemeGroupVersion.WithResource("subnets"):
		object := obj.(*ocicorev1alpha1.Subnet)
		return clientset.OcicoreV1alpha1().Subnets(object.Namespace).Update(object)
	case ocicorev1alpha1.SchemeGroupVersion.WithResource("vcns"):
		object := obj.(*ocicorev1alpha1.Vcn)
		return clientset.OcicoreV1alpha1().Vcns(object.Namespace).Update(object)
	case ocicorev1alpha1.SchemeGroupVersion.WithResource("volumes"):
		object := obj.(*ocicorev1alpha1.Volume)
		return clientset.OcicoreV1alpha1().Volumes(object.Namespace).Update(object)
	case ocilbv1alpha1.SchemeGroupVersion.WithResource("loadbalancers"):
		object := obj.(*ocilbv1alpha1.LoadBalancer)
		return clientset.OcilbV1alpha1().LoadBalancers(object.Namespace).Update(object)
	case ocilbv1alpha1.SchemeGroupVersion.WithResource("backendsets"):
		object := obj.(*ocilbv1alpha1.BackendSet)
		return clientset.OcilbV1alpha1().BackendSets(object.Namespace).Update(object)
	case ocilbv1alpha1.SchemeGroupVersion.WithResource("backends"):
		object := obj.(*ocilbv1alpha1.Backend)
		return clientset.OcilbV1alpha1().Backends(object.Namespace).Update(object)
	case ocilbv1alpha1.SchemeGroupVersion.WithResource("listeners"):
		object := obj.(*ocilbv1alpha1.Listener)
		return clientset.OcilbV1alpha1().Listeners(object.Namespace).Update(object)
	case ocicev1alpha1.SchemeGroupVersion.WithResource("clusters"):
		object := obj.(*ocicev1alpha1.Cluster)
		return clientset.OciceV1alpha1().Clusters(object.Namespace).Update(object)
	case ocicev1alpha1.SchemeGroupVersion.WithResource("nodepools"):
		object := obj.(*ocicev1alpha1.NodePool)
		return clientset.OciceV1alpha1().NodePools(object.Namespace).Update(object)
	}

	return nil, fmt.Errorf("no client found for %v", resource)
}
