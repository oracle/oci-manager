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

package ce

import (
	"context"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"os"
	"reflect"

	ocisdkcommon "github.com/oracle/oci-go-sdk/common"
	ocisdkce "github.com/oracle/oci-go-sdk/containerengine"
	ocisdkcore "github.com/oracle/oci-go-sdk/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"bytes"
	"errors"
	"fmt"
	cegroup "github.com/oracle/oci-manager/pkg/apis/ocice.oracle.com"
	ocicev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocice.oracle.com/v1alpha1"
	ocicommon "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
	"io/ioutil"
	"strings"
)

func init() {
	resourcescommon.RegisterResourceTypeWithValidation(
		cegroup.GroupName,
		ocicev1alpha1.ClusterKind,
		ocicev1alpha1.ClusterResourcePlural,
		ocicev1alpha1.ClusterControllerName,
		&ocicev1alpha1.ClusterValidation,
		NewClusterAdapter)
}

// ClusterAdapter implements the adapter interface for cluster resource
type ClusterAdapter struct {
	clientset versioned.Interface
	ctx       context.Context
	ceClient  resourcescommon.ContainerEngineClientInterface
	vcnClient resourcescommon.VcnClientInterface
}

// NewClusterAdapter creates a new adapter for cluster resource
func NewClusterAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	ca := ClusterAdapter{}
	ca.clientset = clientset
	ca.ctx = context.Background()

	ceClient, err := ocisdkce.NewContainerEngineClientWithConfigurationProvider(ociconfig)
	if err != nil {
		glog.Errorf("Error creating oci ContainerEngine client: %v", err)
		os.Exit(1)
	}
	ca.ceClient = &ceClient

	vcnClient, err := ocisdkcore.NewVirtualNetworkClientWithConfigurationProvider(ociconfig)
	if err != nil {
		glog.Errorf("Error creating oci VCN client: %v", err)
		os.Exit(1)
	}
	ca.vcnClient = &vcnClient
	return &ca
}

// Kind returns the resource kind string
func (a *ClusterAdapter) Kind() string {
	return ocicev1alpha1.ClusterKind
}

// Resource returns the plural name of the resource type
func (a *ClusterAdapter) Resource() string {
	return ocicev1alpha1.ClusterResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *ClusterAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ocicev1alpha1.SchemeGroupVersion.WithResource(ocicev1alpha1.ClusterResourcePlural)
}

// ObjectType returns the cluster type for this adapter
func (a *ClusterAdapter) ObjectType() runtime.Object {
	return &ocicev1alpha1.Cluster{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *ClusterAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ocicev1alpha1.Cluster)
	return ok
}

// Copy returns a copy of a cluster object
func (a *ClusterAdapter) Copy(obj runtime.Object) runtime.Object {
	Cluster := obj.(*ocicev1alpha1.Cluster)
	return Cluster.DeepCopyObject()
}

// Equivalent checks if two cluster objects are the same
func (a *ClusterAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	return true
}

// IsResourceComplient checks if resource config is complient with CRD spec
func (a *ClusterAdapter) IsResourceCompliant(obj runtime.Object) bool {
	cluster := obj.(*ocicev1alpha1.Cluster)

	if cluster.Status.WorkRequestId != nil {
		glog.Infof("cluster has workrequest - isResourceCompliant: false")
		return false
	}

	if cluster.Status.Resource == nil {
		return false
	}

	if cluster.Status.KubeConfig == nil || *cluster.Status.KubeConfig == "" ||
		*cluster.Spec.KubernetesVersion != *cluster.Status.Resource.KubernetesVersion {
		return false
	}

	return true
}

// IsResourceStatusChanged checks if two cluster objects are the same
func (a *ClusterAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {
	cluster1 := obj1.(*ocicev1alpha1.Cluster)
	cluster2 := obj2.(*ocicev1alpha1.Cluster)

	if cluster1.Status.Resource.LifecycleState != cluster2.Status.Resource.LifecycleState {
		return true
	}

	return false
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *ClusterAdapter) Id(obj runtime.Object) string {
	return obj.(*ocicev1alpha1.Cluster).GetResourceID()
}

// ObjectMeta returns the object meta struct from the cluster object
func (a *ClusterAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ocicev1alpha1.Cluster).ObjectMeta
}

// DependsOn returns a map of cluster dependencies (objects that the cluster depends on)
func (a *ClusterAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ocicev1alpha1.Cluster).Spec.DependsOn
}

// Dependents returns a map of cluster dependents (objects that depend on the cluster)
func (a *ClusterAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ocicev1alpha1.Cluster).Status.Dependents
}

// CreateObject creates the cluster object
func (a *ClusterAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicev1alpha1.Cluster)
	return a.clientset.OciceV1alpha1().Clusters(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the cluster object
func (a *ClusterAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicev1alpha1.Cluster)
	return a.clientset.OciceV1alpha1().Clusters(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the cluster object
func (a *ClusterAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var be = obj.(*ocicev1alpha1.Cluster)
	return a.clientset.OciceV1alpha1().Clusters(be.ObjectMeta.Namespace).Delete(be.Name, options)
}

// DependsOnRefs returns the objects that the cluster depends on
func (a *ClusterAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	var cluster = obj.(*ocicev1alpha1.Cluster)
	deps := make([]runtime.Object, 0)

	if !resourcescommon.IsOcid(cluster.Spec.CompartmentRef) {
		c, err := resourcescommon.Compartment(a.clientset, cluster.ObjectMeta.Namespace, cluster.Spec.CompartmentRef)
		if err != nil {
			glog.Errorf("Cluster DependsOnRefs CompartmentRef err: %v", err)
			return nil, err
		}
		deps = append(deps, c)
	}

	if !resourcescommon.IsOcid(cluster.Spec.VcnRef) {
		c, err := resourcescommon.Vcn(a.clientset, cluster.ObjectMeta.Namespace, cluster.Spec.VcnRef)
		if err != nil {
			glog.Errorf("Cluster DependsOnRefs VcnRef err: %v", err)
			return nil, err
		}
		deps = append(deps, c)
	}

	for _, snName := range cluster.Spec.ServiceLbSubnetRefs {
		if !resourcescommon.IsOcid(snName) {
			subnet, err := resourcescommon.Subnet(a.clientset, cluster.ObjectMeta.Namespace, snName)
			if err != nil {
				glog.Errorf("Cluster DependsOnRefs ServiceLbSubnetRefs err: %v", err)
				return nil, err
			}
			deps = append(deps, subnet)
		}
	}
	return deps, nil
}

// Create creates the cluster resource in oci
func (a *ClusterAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	cluster := obj.(*ocicev1alpha1.Cluster)

	if cluster.Status.WorkRequestId != nil {

		workRequest := ocisdkce.GetWorkRequestRequest{WorkRequestId: cluster.Status.WorkRequestId}
		workResp, e := a.ceClient.GetWorkRequest(a.ctx, workRequest)
		if e != nil {
			glog.Errorf("CreateCluster GetWorkRequest error: %v", e)
			return cluster, cluster.Status.HandleError(e)
		}

		glog.V(4).Infof("CreateCluster workResp state: %s", workResp.Status)

		if workResp.Status != ocisdkce.WorkRequestStatusSucceeded &&
			workResp.Status != ocisdkce.WorkRequestStatusFailed {

			if workResp.Status != *cluster.Status.WorkRequestStatus {
				cluster.Status.WorkRequestStatus = &workResp.Status
				return cluster, nil
			} else {
				return nil, nil
			}
		}

		if workResp.Status == ocisdkce.WorkRequestStatusFailed {
			cluster.Status.WorkRequestStatus = &workResp.Status
			err := fmt.Errorf("WorkRequest %s is in failed state", *cluster.Status.WorkRequestId)
			return cluster, cluster.Status.HandleError(err)
		}
		cluster.Status.WorkRequestId = nil
		cluster.Status.WorkRequestStatus = nil
		cluster.Status.Resource = &ocicev1alpha1.ClusterResource{
			Cluster: &ocisdkce.Cluster{
				Id: workResp.Resources[0].Identifier,
			},
		}

	} else {

		compartment, err := resourcescommon.Compartment(a.clientset, cluster.ObjectMeta.Namespace, cluster.Spec.CompartmentRef)
		if err != nil {
			return cluster, cluster.Status.HandleError(err)
		}

		vcn, err := resourcescommon.Vcn(a.clientset, cluster.ObjectMeta.Namespace, cluster.Spec.VcnRef)
		if err != nil {
			return cluster, cluster.Status.HandleError(err)
		}

		subnets := make([]string, 0)
		for _, subnetName := range cluster.Spec.ServiceLbSubnetRefs {
			if resourcescommon.IsOcid(subnetName) {
				subnets = append(subnets, subnetName)
			} else {
				subnetId, err := resourcescommon.SubnetId(a.clientset, cluster.ObjectMeta.Namespace, subnetName)
				if err != nil {
					return cluster, cluster.Status.HandleError(err)
				}
				subnets = append(subnets, subnetId)
			}
		}
		glog.V(1).Infof("CreateCluster Service LB subnets: %s", subnets)
		cluster.Spec.Options.ServiceLbSubnetIds = subnets

		details := ocisdkce.CreateClusterDetails{
			Name:              &cluster.ObjectMeta.Name,
			Options:           cluster.Spec.Options,
			VcnId:             vcn.Status.Resource.Id,
			KubernetesVersion: cluster.Spec.KubernetesVersion,
			CompartmentId:     compartment.Status.Resource.Id,
		}

		createRequest := ocisdkce.CreateClusterRequest{
			CreateClusterDetails: details,
			OpcRetryToken:        ocisdkcommon.String(string(cluster.UID)),
		}

		glog.V(1).Infof("CreateCluster: %s OpcRetryToken: %s", cluster.ObjectMeta.Name, string(cluster.UID))

		createResponse, e := a.ceClient.CreateCluster(a.ctx, createRequest)

		if e != nil {
			glog.Errorf("CreateCluster error: %v", e)
			return cluster, cluster.Status.HandleError(e)
		} else {
			workRequestId := *createResponse.OpcWorkRequestId
			glog.V(4).Infof("CreateCluster workRequestId: %s", workRequestId)
			cluster.Status.WorkRequestId = createResponse.OpcWorkRequestId
			stateAccepted := ocisdkce.WorkRequestStatusAccepted
			cluster.Status.WorkRequestStatus = &stateAccepted
			return cluster, nil

		}
	}

	return a.getCluster(cluster)
}

// Delete deletes the cluster resource in oci
func (a *ClusterAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	var cluster = obj.(*ocicev1alpha1.Cluster)
	if cluster.Status.Resource != nil && cluster.Status.Resource.Id != nil {
		glog.Infof("DeleteCluster - clusterId: %s", *cluster.Status.Resource.Id)
	} else {
		return nil, errors.New(fmt.Sprintf("missing clusterId. cluster: %v", cluster))
	}

	deleteRequest := ocisdkce.DeleteClusterRequest{
		ClusterId: cluster.Status.Resource.Id,
	}
	_, err := a.ceClient.DeleteCluster(a.ctx, deleteRequest)
	if err != nil {
		glog.Errorf("DeleteCluster name: %s error: %v", cluster.Name, err)
		return nil, cluster.Status.HandleError(err)
	}
	glog.Infof("DeleteCluster: %s ok", cluster.Name)
	return cluster, nil
}

// Get retrieves the cluster resource from oci
func (a *ClusterAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	var cluster = obj.(*ocicev1alpha1.Cluster)

	getClusterReq := ocisdkce.GetClusterRequest{
		ClusterId: cluster.Status.Resource.Id,
	}
	clusterResp, e := a.ceClient.GetCluster(a.ctx, getClusterReq)

	if e != nil {
		return cluster, cluster.Status.HandleError(e)
	}

	return cluster.SetResource(&clusterResp.Cluster), nil
}

// Update updates the cluster resource in oci
func (a *ClusterAdapter) Update(obj runtime.Object) (runtime.Object, error) {

	cluster := obj.(*ocicev1alpha1.Cluster)

	glog.V(2).Infof("UpdateCluster: %s", cluster.Name)

	if cluster.Status.WorkRequestId != nil {

		workRequest := ocisdkce.GetWorkRequestRequest{WorkRequestId: cluster.Status.WorkRequestId}
		workResp, e := a.ceClient.GetWorkRequest(a.ctx, workRequest)
		if e != nil {
			return cluster, cluster.Status.HandleError(e)
		}
		cluster.Status.WorkRequestStatus = &workResp.Status

		if workResp.Status != ocisdkce.WorkRequestStatusFailed &&
			workResp.Status != ocisdkce.WorkRequestStatusSucceeded {

			if workResp.Status != *cluster.Status.WorkRequestStatus {
				cluster.Status.WorkRequestStatus = &workResp.Status
				return cluster, nil
			} else {
				return nil, nil
			}
		}

		if workResp.Status == ocisdkce.WorkRequestStatusFailed {
			cluster.Status.WorkRequestStatus = &workResp.Status
			err := fmt.Errorf("WorkRequest %s is in failed state", *cluster.Status.WorkRequestId)
			return cluster, cluster.Status.HandleError(err)
		}

		cluster.Status.WorkRequestId = nil
		cluster.Status.WorkRequestStatus = nil
		cluster.Status.Resource = &ocicev1alpha1.ClusterResource{
			Cluster: &ocisdkce.Cluster{
				Id: workResp.Resources[0].Identifier,
			},
		}

	} else {
		details := ocisdkce.UpdateClusterDetails{
			KubernetesVersion: cluster.Spec.KubernetesVersion,
		}

		updateRequest := ocisdkce.UpdateClusterRequest{
			ClusterId:            cluster.Status.Resource.Id,
			UpdateClusterDetails: details,
		}

		updateResponse, e := a.ceClient.UpdateCluster(a.ctx, updateRequest)
		if e != nil && !strings.Contains(e.Error(), "must be different than") {
			glog.Errorf("UpdateCluster error: %s", e)
			return cluster, cluster.Status.HandleError(e)
		}
		if e == nil {
			workRequestId := *updateResponse.OpcWorkRequestId
			glog.Infof("UpdateCluster workRequestId: %s", workRequestId)

			workRequest := ocisdkce.GetWorkRequestRequest{WorkRequestId: &workRequestId}
			workResp, e := a.ceClient.GetWorkRequest(a.ctx, workRequest)
			if e != nil {
				return cluster, cluster.Status.HandleError(e)
			}

			cluster.Status.WorkRequestStatus = &workResp.Status
			cluster.Status.WorkRequestId = updateResponse.OpcWorkRequestId
			return cluster, nil
		}

	}

	return a.getCluster(cluster)
}

// populates the cluster status resource and kubeconfig
func (a *ClusterAdapter) getCluster(cluster *ocicev1alpha1.Cluster) (runtime.Object, error) {

	getClusterReq := ocisdkce.GetClusterRequest{
		ClusterId: cluster.Status.Resource.Id,
	}

	clusterResp, e := a.ceClient.GetCluster(a.ctx, getClusterReq)
	if e != nil {
		glog.Errorf("GetCluster error: %v", e)
		return cluster, cluster.Status.HandleError(e)
	}

	createKubeconfigReq := ocisdkce.CreateKubeconfigRequest{
		ClusterId: cluster.Status.Resource.Id,
	}
	kubeConfigResp, e := a.ceClient.CreateKubeconfig(a.ctx, createKubeconfigReq)
	if e != nil {
		glog.Errorf("Kubeconfig error: %v", e)
		return cluster, cluster.Status.HandleError(e)
	}

	// get kubeconfig
	var bodyBytes []byte
	if kubeConfigResp.Content != nil {
		bodyBytes, _ = ioutil.ReadAll(kubeConfigResp.Content)
	}
	// Restore the io.ReadCloser to its original state
	kubeConfigResp.Content = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	// Use the content
	kubeconfig := string(bodyBytes)
	if !reflect.DeepEqual(cluster.Status.KubeConfig, kubeconfig) {
		cluster.Status.KubeConfig = &kubeconfig
	}

	return cluster.SetResource(&clusterResp.Cluster), cluster.Status.HandleError(e)
}

// UpdateForResource calls a common UpdateForResource method to update the cluster resource in the cluster object
func (a *ClusterAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
