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

package fake

import (
	"context"
	"io/ioutil"

	ocice "github.com/oracle/oci-go-sdk/containerengine"

	"k8s.io/apimachinery/pkg/util/uuid"

	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"

	"bytes"
)

const (
	// FakeWorkRequestID used as work request id for fake responses
	FakeCeWorkRequestID = "wrID123"
)

var (
	// fakeContainerEngineID used as cluster id for fake responses
	fakeClusterID = "clusterID123"
)

// ContainerEngineClient implements common ContainerEngineClient to fake oci methods for unit tests.
type ContainerEngineClient struct {
	resourcescommon.ContainerEngineClientInterface

	clusters  map[string]ocice.Cluster
	nodepools map[string]ocice.NodePool

	workRequests map[string]ocice.WorkRequest
}

// NewContainerEngineClient returns a client that will respond with the provided objects.
// It shouldn't be used as a replacement for a real client and is mostly useful in simple unit tests.
func NewContainerEngineClient() (fcec *ContainerEngineClient) {
	c := ContainerEngineClient{}
	c.clusters = make(map[string]ocice.Cluster)
	c.nodepools = make(map[string]ocice.NodePool)

	c.workRequests = make(map[string]ocice.WorkRequest)
	return &c
}

// CreateCluster returns a fake response
func (cec *ContainerEngineClient) CreateCluster(ctx context.Context, request ocice.CreateClusterRequest) (response ocice.CreateClusterResponse, err error) {

	response = ocice.CreateClusterResponse{}
	wrID := FakeCeWorkRequestID
	response.OpcWorkRequestId = &wrID

	cluster := ocice.Cluster{}
	cluster.Name = request.Name

	cec.clusters[fakeClusterID] = cluster

	return response, nil
}

// DeleteCluster returns a fake response
func (cec *ContainerEngineClient) DeleteCluster(ctx context.Context, request ocice.DeleteClusterRequest) (response ocice.DeleteClusterResponse, err error) {
	cID := *request.ClusterId
	delete(cec.clusters, cID)

	response = ocice.DeleteClusterResponse{}
	reqid := string(uuid.NewUUID())
	response.OpcRequestId = &reqid
	return response, nil
}

func (cec *ContainerEngineClient) GetWorkRequest(ctx context.Context, request ocice.GetWorkRequestRequest) (response ocice.GetWorkRequestResponse, err error) {

	response = ocice.GetWorkRequestResponse{
		WorkRequest: ocice.WorkRequest{
			Status:    ocice.WorkRequestStatusSucceeded,
			Resources: []ocice.WorkRequestResource{{Identifier: &fakeClusterID}},
		},
	}

	return response, nil // servicefailure{Message: "Not found", Code: "NotAuthorizedOrNotFound"}
}

// GetCluster returns a fake response
func (cec *ContainerEngineClient) GetCluster(ctx context.Context, request ocice.GetClusterRequest) (response ocice.GetClusterResponse, err error) {
	response = ocice.GetClusterResponse{
		Cluster: cec.clusters[fakeClusterID],
	}
	return response, nil
}

// GetCluster returns a fake response
func (cec *ContainerEngineClient) UpdateCluster(ctx context.Context, request ocice.UpdateClusterRequest) (response ocice.UpdateClusterResponse, err error) {
	response = ocice.UpdateClusterResponse{}
	return response, nil
}

func (cec *ContainerEngineClient) CreateKubeconfig(ctx context.Context, request ocice.CreateKubeconfigRequest) (response ocice.CreateKubeconfigResponse, err error) {

	response = ocice.CreateKubeconfigResponse{
		Content: ioutil.NopCloser(bytes.NewReader([]byte(""))),
	}
	return response, nil
}

// CreateNodePool returns a fake response
func (cec *ContainerEngineClient) CreateNodePool(ctx context.Context, request ocice.CreateNodePoolRequest) (response ocice.CreateNodePoolResponse, err error) {

	np := ocice.NodePool{
		NodeShape:         request.NodeShape,
		NodeImageName:     request.NodeImageName,
		QuantityPerSubnet: request.QuantityPerSubnet,
		KubernetesVersion: request.KubernetesVersion,
		Id:                &fakeClusterID,
		ClusterId:         &fakeClusterID,
		Name:              request.Name,
		CompartmentId:     request.CompartmentId,
	}
	np.Name = request.Name
	cec.nodepools[fakeClusterID] = np
	wrId := FakeWorkRequestID
	response.OpcWorkRequestId = &wrId

	return response, nil
}

// DeleteNodePool returns a fake response
func (cec *ContainerEngineClient) DeleteNodePool(ctx context.Context, request ocice.DeleteNodePoolRequest) (response ocice.DeleteNodePoolResponse, err error) {
	response = ocice.DeleteNodePoolResponse{}

	npID := *request.NodePoolId
	delete(cec.nodepools, npID)

	return response, nil
}

// GetNodePool returns a fake response
func (cec *ContainerEngineClient) GetNodePool(ctx context.Context, request ocice.GetNodePoolRequest) (response ocice.GetNodePoolResponse, err error) {

	response = ocice.GetNodePoolResponse{
		NodePool: cec.nodepools[fakeClusterID],
	}
	return response, nil
}

// UpdateNodePool returns a fake response
func (cec *ContainerEngineClient) UpdateNodePool(ctx context.Context, request ocice.UpdateNodePoolRequest) (response ocice.UpdateNodePoolResponse, err error) {

	wrId := FakeWorkRequestID
	response = ocice.UpdateNodePoolResponse{
		OpcWorkRequestId: &wrId,
	}
	return response, nil
}
