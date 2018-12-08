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
	"errors"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"os"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	ocisdkcommon "github.com/oracle/oci-go-sdk/common"
	ocisdkce "github.com/oracle/oci-go-sdk/containerengine"
	ocisdkcore "github.com/oracle/oci-go-sdk/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"fmt"
	cegroup "github.com/oracle/oci-manager/pkg/apis/ocice.oracle.com"
	ocicev1alpha1 "github.com/oracle/oci-manager/pkg/apis/ocice.oracle.com/v1alpha1"
	ocicommon "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	"github.com/oracle/oci-manager/pkg/client/clientset/versioned"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
	"strings"
)

func init() {
	resourcescommon.RegisterResourceTypeWithValidation(
		cegroup.GroupName,
		ocicev1alpha1.NodePoolKind,
		ocicev1alpha1.NodePoolResourcePlural,
		ocicev1alpha1.NodePoolControllerName,
		&ocicev1alpha1.NodePoolValidation,
		NewNodePoolAdapter)
}

// NodePoolAdapter implements the adapter interface for nodePool resource
type NodePoolAdapter struct {
	clientset versioned.Interface
	ctx       context.Context
	ceClient  resourcescommon.ContainerEngineClientInterface
	vcnClient resourcescommon.VcnClientInterface
}

// NewNodePoolAdapter creates a new adapter for nodePool resource
func NewNodePoolAdapter(clientset versioned.Interface, kubeclient kubernetes.Interface,
	ociconfig ocisdkcommon.ConfigurationProvider, adapterSpecificArgs map[string]interface{}) resourcescommon.ResourceTypeAdapter {
	ca := NodePoolAdapter{}
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
func (a *NodePoolAdapter) Kind() string {
	return ocicev1alpha1.NodePoolKind
}

// Resource returns the plural name of the resource type
func (a *NodePoolAdapter) Resource() string {
	return ocicev1alpha1.NodePoolResourcePlural
}

// GroupVersionWithResource returns the group version schema with the resource type
func (a *NodePoolAdapter) GroupVersionWithResource() schema.GroupVersionResource {
	return ocicev1alpha1.SchemeGroupVersion.WithResource(ocicev1alpha1.NodePoolResourcePlural)
}

// ObjectType returns the nodePool type for this adapter
func (a *NodePoolAdapter) ObjectType() runtime.Object {
	return &ocicev1alpha1.NodePool{}
}

// IsExpectedType ensures the resource type matches the adapter type
func (a *NodePoolAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*ocicev1alpha1.NodePool)
	return ok
}

// Copy returns a copy of a nodePool object
func (a *NodePoolAdapter) Copy(obj runtime.Object) runtime.Object {
	NodePool := obj.(*ocicev1alpha1.NodePool)
	return NodePool.DeepCopyObject()
}

// Equivalent checks if two nodePool objects are the same
func (a *NodePoolAdapter) Equivalent(obj1, obj2 runtime.Object) bool {
	return true
}

// IsResourceCompliant checks if resource config is complient with CRD spec
func (a *NodePoolAdapter) IsResourceCompliant(obj runtime.Object) bool {
	nodePool := obj.(*ocicev1alpha1.NodePool)

	if nodePool.Status.WorkRequestStatus != nil {
		glog.Infof("nodePool has workrequest - isResourceCompliant: false")
		return false
	}

	if nodePool.Status.Resource == nil {
		return false
	}

	if *nodePool.Spec.KubernetesVersion != *nodePool.Status.Resource.KubernetesVersion ||
		*nodePool.Spec.QuantityPerSubnet != *nodePool.Status.Resource.QuantityPerSubnet {
		return false
	}

	return true
}

// IsResourceStatusChanged checks if two vcn objects are the same
func (a *NodePoolAdapter) IsResourceStatusChanged(obj1, obj2 runtime.Object) bool {

	np1 := obj1.(*ocicev1alpha1.NodePool)
	np2 := obj2.(*ocicev1alpha1.NodePool)

	if (np1.Status.WorkRequestId == nil && np2.Status.WorkRequestId != nil) ||
		(np1.Status.WorkRequestId != nil && np2.Status.WorkRequestId == nil) {
		return true
	}

	return false
}

// Id returns the unique resource id via the object type method (i.e the oci id)
func (a *NodePoolAdapter) Id(obj runtime.Object) string {
	return obj.(*ocicev1alpha1.NodePool).GetResourceID()
}

// ObjectMeta returns the object meta struct from the nodePool object
func (a *NodePoolAdapter) ObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	return &obj.(*ocicev1alpha1.NodePool).ObjectMeta
}

// DependsOn returns a map of nodePool dependencies (objects that the nodePool depends on)
func (a *NodePoolAdapter) DependsOn(obj runtime.Object) map[string]ocicommon.DependsOn {
	return obj.(*ocicev1alpha1.NodePool).Spec.DependsOn
}

// Dependents returns a map of nodePool dependents (objects that depend on the nodePool)
func (a *NodePoolAdapter) Dependents(obj runtime.Object) map[string][]string {
	return obj.(*ocicev1alpha1.NodePool).Status.Dependents
}

// CreateObject creates the nodePool object
func (a *NodePoolAdapter) CreateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicev1alpha1.NodePool)
	return a.clientset.OciceV1alpha1().NodePools(object.ObjectMeta.Namespace).Create(object)
}

// UpdateObject updates the nodePool object
func (a *NodePoolAdapter) UpdateObject(obj runtime.Object) (runtime.Object, error) {
	var object = obj.(*ocicev1alpha1.NodePool)
	return a.clientset.OciceV1alpha1().NodePools(object.ObjectMeta.Namespace).Update(object)
}

// DeleteObject deletes the nodePool object
func (a *NodePoolAdapter) DeleteObject(obj runtime.Object, options *metav1.DeleteOptions) error {
	var be = obj.(*ocicev1alpha1.NodePool)
	return a.clientset.OciceV1alpha1().NodePools(be.ObjectMeta.Namespace).Delete(be.Name, options)
}

// DependsOnRefs returns the objects that the nodePool depends on
func (a *NodePoolAdapter) DependsOnRefs(obj runtime.Object) ([]runtime.Object, error) {
	var nodePool = obj.(*ocicev1alpha1.NodePool)
	deps := make([]runtime.Object, 0)

	if !resourcescommon.IsOcid(nodePool.Spec.CompartmentRef) {
		c, err := resourcescommon.Compartment(a.clientset, nodePool.ObjectMeta.Namespace, nodePool.Spec.CompartmentRef)
		if err != nil {
			glog.Errorf("NodePool DependsOnRefs CompartmentRef err: %v", err)
			return nil, err
		}
		deps = append(deps, c)
	}

	for _, snName := range nodePool.Spec.SubnetRefs {
		if !resourcescommon.IsOcid(snName) {
			subnet, err := resourcescommon.Subnet(a.clientset, nodePool.ObjectMeta.Namespace, snName)
			if err != nil {
				glog.Errorf("NodePool DependsOnRefs SubnetRefs err: %v", err)
				return nil, err
			}
			deps = append(deps, subnet)
		}
	}

	if !resourcescommon.IsOcid(nodePool.Spec.ClusterRef) {
		c, err := resourcescommon.Cluster(a.clientset, nodePool.ObjectMeta.Namespace, nodePool.Spec.ClusterRef)
		if err != nil {
			glog.Errorf("NodePool DependsOnRefs ClusterRef err: %v", err)
			return nil, err
		}
		deps = append(deps, c)
	}

	return deps, nil
}

// Create creates the nodePool resource in oci
func (a *NodePoolAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	nodePool := obj.(*ocicev1alpha1.NodePool)

	if nodePool.Status.WorkRequestId != nil {

		workRequest := ocisdkce.GetWorkRequestRequest{WorkRequestId: nodePool.Status.WorkRequestId}
		workResp, e := a.ceClient.GetWorkRequest(a.ctx, workRequest)
		if e != nil {
			glog.Errorf("CreateNodePool GetWorkRequest error: %v", e)
			return nodePool, nodePool.Status.HandleError(e)
		}
		glog.Infof("CreateNodePool workResp state: %s", workResp.Status)

		if workResp.Status != ocisdkce.WorkRequestStatusSucceeded &&
			workResp.Status != ocisdkce.WorkRequestStatusFailed {

			if workResp.Status != *nodePool.Status.WorkRequestStatus {
				nodePool.Status.WorkRequestStatus = &workResp.Status
				return nodePool, nil

			} else {
				return nil, nil
			}

		}
		if workResp.Status == ocisdkce.WorkRequestStatusFailed {
			nodePool.Status.WorkRequestStatus = &workResp.Status
			err := fmt.Errorf("WorkRequest %s is in failed state", *nodePool.Status.WorkRequestId)
			return nodePool, nodePool.Status.HandleError(err)
		}

		nodePool.Status.WorkRequestId = nil
		nodePool.Status.WorkRequestStatus = nil
		nodePool.Status.Resource = &ocicev1alpha1.NodePoolResource{
			NodePool: &ocisdkce.NodePool{
				Id: workResp.Resources[0].Identifier,
			},
		}

	} else {

		compartment, err := resourcescommon.Compartment(a.clientset, nodePool.ObjectMeta.Namespace, nodePool.Spec.CompartmentRef)
		if err != nil {
			return nodePool, nodePool.Status.HandleError(err)
		}

		cluster, err := resourcescommon.Cluster(a.clientset, nodePool.ObjectMeta.Namespace, nodePool.Spec.ClusterRef)
		if err != nil {
			return nodePool, nodePool.Status.HandleError(err)
		}

		subnets := make([]string, 0)
		for _, subnetName := range nodePool.Spec.SubnetRefs {
			if resourcescommon.IsOcid(subnetName) {
				subnets = append(subnets, subnetName)
			} else {
				subnetId, err := resourcescommon.SubnetId(a.clientset, nodePool.ObjectMeta.Namespace, subnetName)
				if err != nil {
					return nodePool, nodePool.Status.HandleError(err)
				}
				subnets = append(subnets, subnetId)
			}
		}

		if cluster.Status.Resource != nil && cluster.Status.Resource.Id != nil {
			glog.Infof("CreateNodePool Cluster - clusterId: %s", *cluster.Status.Resource.Id)
		} else {
			return nil, errors.New(fmt.Sprintf("missing cluster resource clusterId. cluster: %v", cluster))
		}

		details := ocisdkce.CreateNodePoolDetails{
			Name:              &nodePool.ObjectMeta.Name,
			ClusterId:         cluster.Status.Resource.Id,
			InitialNodeLabels: nodePool.Spec.InitialNodeLabels,
			KubernetesVersion: nodePool.Spec.KubernetesVersion,
			NodeImageName:     nodePool.Spec.NodeImageName,
			NodeShape:         nodePool.Spec.NodeShape,
			QuantityPerSubnet: nodePool.Spec.QuantityPerSubnet,
			SshPublicKey:      nodePool.Spec.SshPublicKey,
			SubnetIds:         subnets,

			CompartmentId: compartment.Status.Resource.CompartmentId,
		}

		createRequest := ocisdkce.CreateNodePoolRequest{
			CreateNodePoolDetails: details,
			OpcRetryToken:         ocisdkcommon.String(string(nodePool.UID)),
		}
		glog.V(4).Infof("CreateNodePool %v", details)
		glog.V(4).Infof("NodePool: %s OpcRetryToken: %s", nodePool.ObjectMeta.Name, nodePool.UID)
		createResponse, e := a.ceClient.CreateNodePool(a.ctx, createRequest)
		if e != nil {
			glog.Errorf("CreateNodePool error: %v", e)
			return nodePool, nodePool.Status.HandleError(e)
		}
		workRequestId := *createResponse.OpcWorkRequestId
		glog.V(4).Infof("CreateNodePool workRequestId: %s", workRequestId)

		workRequest := ocisdkce.GetWorkRequestRequest{WorkRequestId: &workRequestId}
		workResp, e := a.ceClient.GetWorkRequest(a.ctx, workRequest)
		if e != nil {
			glog.Errorf("CreateNodePool GetWorkRequest error: %v", e)
			return nodePool, nodePool.Status.HandleError(e)
		}
		glog.V(4).Infof("CreateNodePool workResp state: %s", workResp.Status)

		nodePool.Status.WorkRequestId = createResponse.OpcWorkRequestId
		nodePool.Status.WorkRequestStatus = &workResp.Status

		return nodePool, nil
	}

	return a.Get(nodePool)
}

// Delete deletes the nodePool resource in oci
func (a *NodePoolAdapter) Delete(obj runtime.Object) (runtime.Object, error) {
	var np = obj.(*ocicev1alpha1.NodePool)

	deleteRequest := ocisdkce.DeleteNodePoolRequest{
		NodePoolId: np.Status.Resource.Id,
	}
	_, err := a.ceClient.DeleteNodePool(a.ctx, deleteRequest)
	if err != nil {
		if strings.Contains(err.Error(), "IncorrectState") {
			getRequest := ocisdkce.GetNodePoolRequest{
				NodePoolId: np.Status.Resource.Id,
			}

			getResp, err := a.ceClient.GetNodePool(a.ctx, getRequest)
			if err != nil && apierrors.IsNotFound(err) || getResp.NodePool.Id == nil {
				return np, nil
			} else {
				glog.Errorf("get nodepool in delete: %v", getResp)
			}

		}
		glog.Errorf("DeleteNodePool name: %s error: %v", np.Name, err)
	}
	glog.Infof("DeleteNodePool: %s ok", np.Name)

	return np, np.Status.HandleError(err)
}

// Get retrieves the nodePool resource from oci
func (a *NodePoolAdapter) Get(obj runtime.Object) (runtime.Object, error) {
	nodePool := obj.(*ocicev1alpha1.NodePool)

	getNodePoolReq := ocisdkce.GetNodePoolRequest{
		NodePoolId: nodePool.Status.Resource.Id,
	}
	nodePoolResp, e := a.ceClient.GetNodePool(a.ctx, getNodePoolReq)
	if e != nil {
		return nodePool, nodePool.Status.HandleError(e)
	}

	return nodePool.SetResource(&nodePoolResp.NodePool), nodePool.Status.HandleError(e)
}

// Update updates the nodePool resource in oci
func (a *NodePoolAdapter) Update(obj runtime.Object) (runtime.Object, error) {
	var nodePool = obj.(*ocicev1alpha1.NodePool).DeepCopy()

	if nodePool.Status.WorkRequestId != nil {

		workRequest := ocisdkce.GetWorkRequestRequest{WorkRequestId: nodePool.Status.WorkRequestId}
		workResp, e := a.ceClient.GetWorkRequest(a.ctx, workRequest)
		if e != nil {
			return nodePool, nodePool.Status.HandleError(e)
		}
		nodePool.Status.WorkRequestStatus = &workResp.Status

		if workResp.Status != ocisdkce.WorkRequestStatusFailed &&
			workResp.Status != ocisdkce.WorkRequestStatusSucceeded {

			if workResp.Status != *nodePool.Status.WorkRequestStatus {
				nodePool.Status.WorkRequestStatus = &workResp.Status
				return nodePool, nil
			} else {
				return nil, nil
			}
		}

		if workResp.Status == ocisdkce.WorkRequestStatusFailed {
			nodePool.Status.WorkRequestStatus = &workResp.Status
			err := fmt.Errorf("WorkRequest %s is in failed state", *nodePool.Status.WorkRequestId)
			return nodePool, nodePool.Status.HandleError(err)
		}

		nodePool.Status.WorkRequestId = nil
		nodePool.Status.WorkRequestStatus = nil

	} else {

		if nodePool.Status.Resource != nil && nodePool.Status.Resource.Id != nil {
			glog.Infof("UpdateCluster - clusterId: %s", *nodePool.Status.Resource.Id)
		} else {
			return nil, errors.New(fmt.Sprintf("missing nodePoolId. nodePool: %v", nodePool))
		}

		details := ocisdkce.UpdateNodePoolDetails{
			Name:              &nodePool.ObjectMeta.Name,
			KubernetesVersion: nodePool.Spec.KubernetesVersion,
			QuantityPerSubnet: nodePool.Spec.QuantityPerSubnet,
		}

		updateRequest := ocisdkce.UpdateNodePoolRequest{
			NodePoolId:            nodePool.Status.Resource.Id,
			UpdateNodePoolDetails: details,
		}

		updateResponse, e := a.ceClient.UpdateNodePool(a.ctx, updateRequest)
		if e != nil {
			glog.Errorf("UpdateNodePool error: %v", e)
			return nodePool, nodePool.Status.HandleError(e)
		}
		workRequestId := *updateResponse.OpcWorkRequestId
		glog.Infof("UpdateNodePool workRequestId: %s", workRequestId)

		workRequest := ocisdkce.GetWorkRequestRequest{WorkRequestId: &workRequestId}
		workResp, e := a.ceClient.GetWorkRequest(a.ctx, workRequest)
		if e != nil {
			glog.Errorf("UpdateNodePool GetWorkRequest error: %v", e)
			return nodePool, nodePool.Status.HandleError(e)
		}
		glog.Infof("UpdateNodePool workResp state: %s", workResp.Status)
		nodePool.Status.WorkRequestId = updateResponse.OpcWorkRequestId
		nodePool.Status.WorkRequestStatus = &workResp.Status
		return nodePool, nil
	}

	return a.Get(nodePool)
}

// UpdateForResource calls a common UpdateForResource method to update the nodePool resource in the nodePool object
func (a *NodePoolAdapter) UpdateForResource(resource schema.GroupVersionResource, obj runtime.Object) (runtime.Object, error) {
	return resourcescommon.UpdateForResource(a.clientset, resource, obj)
}
