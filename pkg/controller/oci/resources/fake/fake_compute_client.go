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

	ocicore "github.com/oracle/oci-go-sdk/core"
	"k8s.io/apimachinery/pkg/util/uuid"

	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
)

// ComputeClient implements common ComputeClient to fake oci methods for unit tests.
type ComputeClient struct {
	resourcescommon.ComputeClientInterface

	bootVolumeAttaments map[string]ocicore.BootVolumeAttachment
	instances           map[string]ocicore.Instance
	vnicAttachements    map[string]ocicore.VnicAttachment
	volumeAttachements  map[string]ocicore.VolumeAttachment
}

// NewComputeClient returns a client that will respond with the provided objects.
// It shouldn't be used as a replacement for a real client and is mostly useful in simple unit tests.
func NewComputeClient() (fcc *ComputeClient) {
	c := ComputeClient{}
	c.bootVolumeAttaments = make(map[string]ocicore.BootVolumeAttachment)
	c.instances = make(map[string]ocicore.Instance)
	c.vnicAttachements = make(map[string]ocicore.VnicAttachment)
	c.volumeAttachements = make(map[string]ocicore.VolumeAttachment)
	return &c
}

// AttachVolume returns a fake response for AttachVolume
func (cc *ComputeClient) AttachVolume(ctx context.Context, request ocicore.AttachVolumeRequest) (response ocicore.AttachVolumeResponse, err error) {
	return response, nil
}

// DetachVolume returns a fake response for DetachVolume
func (cc *ComputeClient) DetachVolume(ctx context.Context, request ocicore.DetachVolumeRequest) (response ocicore.DetachVolumeResponse, err error) {
	return response, nil
}

// GetInstance returns a fake response for GetInstance
func (cc *ComputeClient) GetInstance(ctx context.Context, request ocicore.GetInstanceRequest) (response ocicore.GetInstanceResponse, err error) {
	if instance, ok := cc.instances[*request.InstanceId]; ok {
		response = ocicore.GetInstanceResponse{}
		response.Instance = instance
		return response, nil
	}
	response = ocicore.GetInstanceResponse{}
	return response, servicefailure{Message: "Not found", Code: "NotAuthorizedOrNotFound"}
}

// GetVolumeAttachment returns a fake response for GetVolumeAttachment
func (cc *ComputeClient) GetVolumeAttachment(ctx context.Context, request ocicore.GetVolumeAttachmentRequest) (response ocicore.GetVolumeAttachmentResponse, err error) {
	if va, ok := cc.volumeAttachements[*request.VolumeAttachmentId]; ok {
		response = ocicore.GetVolumeAttachmentResponse{}
		response.VolumeAttachment = va
		return response, nil
	}
	response = ocicore.GetVolumeAttachmentResponse{}
	return response, servicefailure{Message: "Not found", Code: "NotAuthorizedOrNotFound"}
}

// InstanceAction returns a fake response for InstanceAction
func (cc *ComputeClient) InstanceAction(ctx context.Context, request ocicore.InstanceActionRequest) (response ocicore.InstanceActionResponse, err error) {
	response = ocicore.InstanceActionResponse{}
	instance := ocicore.Instance{}
	ocid := string(uuid.NewUUID())
	instance.Id = &ocid
	response.Instance = instance
	cc.instances[ocid] = instance
	return response, nil
}

// LaunchInstance returns a fake response for LaunchInstance
func (cc *ComputeClient) LaunchInstance(ctx context.Context, request ocicore.LaunchInstanceRequest) (response ocicore.LaunchInstanceResponse, err error) {
	response = ocicore.LaunchInstanceResponse{}
	instance := ocicore.Instance{}
	ocid := string(uuid.NewUUID())
	instance.Id = &ocid
	response.Instance = instance
	cc.instances[ocid] = instance
	return response, nil
}

// ListBootVolumeAttachments returns a fake response for ListBootVolumeAttachments
func (cc *ComputeClient) ListBootVolumeAttachments(ctx context.Context, request ocicore.ListBootVolumeAttachmentsRequest) (response ocicore.ListBootVolumeAttachmentsResponse, err error) {
	response = ocicore.ListBootVolumeAttachmentsResponse{}
	reqid := string(uuid.NewUUID())

	n := "bla"
	item := ocicore.BootVolumeAttachment{
		DisplayName:    &n,
		Id:             &n,
		BootVolumeId:   &n,
		LifecycleState: ocicore.BootVolumeAttachmentLifecycleStateAttached,
	}
	response.Items = append(response.Items, item)
	response.OpcRequestId = &reqid
	return response, nil
}

// ListVnicAttachments returns a fake response for ListVnicAttachments
func (cc *ComputeClient) ListVnicAttachments(ctx context.Context, request ocicore.ListVnicAttachmentsRequest) (response ocicore.ListVnicAttachmentsResponse, err error) {
	response = ocicore.ListVnicAttachmentsResponse{}
	reqid := string(uuid.NewUUID())

	n := "bla"
	item := ocicore.VnicAttachment{
		DisplayName:    &n,
		Id:             &n,
		LifecycleState: ocicore.VnicAttachmentLifecycleStateAttached,
	}
	response.Items = append(response.Items, item)
	response.OpcRequestId = &reqid
	return response, nil
}

// ListImages returns a fake response for ListImages
func (cc *ComputeClient) ListImages(ctx context.Context, request ocicore.ListImagesRequest) (response ocicore.ListImagesResponse, err error) {
	response = ocicore.ListImagesResponse{}
	reqid := string(uuid.NewUUID())
	image := "bla"
	item := ocicore.Image{
		Id:          &image,
		DisplayName: &image,
	}
	response.Items = append(response.Items, item)
	response.OpcRequestId = &reqid
	return response, nil
}

// ListShapes returns a fake response for ListShapes
func (cc *ComputeClient) ListShapes(ctx context.Context, request ocicore.ListShapesRequest) (response ocicore.ListShapesResponse, err error) {
	response = ocicore.ListShapesResponse{}
	reqid := string(uuid.NewUUID())
	shape := "bla"
	item := ocicore.Shape{
		Shape: &shape,
	}
	response.Items = append(response.Items, item)
	response.OpcRequestId = &reqid
	return response, nil
}

// TerminateInstance returns a fake response for TerminateInstance
func (cc *ComputeClient) TerminateInstance(ctx context.Context, request ocicore.TerminateInstanceRequest) (response ocicore.TerminateInstanceResponse, err error) {
	id := *request.InstanceId
	delete(cc.instances, id)
	response = ocicore.TerminateInstanceResponse{}
	reqid := string(uuid.NewUUID())
	response.OpcRequestId = &reqid
	return response, nil
}

// UpdateInstance returns a fake response for UpdateInstance
func (cc *ComputeClient) UpdateInstance(ctx context.Context, request ocicore.UpdateInstanceRequest) (response ocicore.UpdateInstanceResponse, err error) {
	if instance, ok := cc.instances[*request.InstanceId]; ok {
		response = ocicore.UpdateInstanceResponse{}
		response.Instance = instance
		return response, nil
	}
	response = ocicore.UpdateInstanceResponse{}
	return response, servicefailure{Message: "Not found", Code: "NotAuthorizedOrNotFound"}
}
