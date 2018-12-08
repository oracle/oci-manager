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

// BlockStorageClient implements common BlockStorageClient to fake oci methods for unit tests.
type BlockStorageClient struct {
	resourcescommon.BlockStorageClientInterface

	backupVolumes map[string]ocicore.VolumeBackup
	bootVolumes   map[string]ocicore.BootVolume
	volumes       map[string]ocicore.Volume
}

// NewBlockStorageClient returns a client that will respond with the provided objects.
// It shouldn't be used as a replacement for a real client and is mostly useful in simple unit tests.
func NewBlockStorageClient() (fcc *BlockStorageClient) {
	c := BlockStorageClient{}
	c.backupVolumes = make(map[string]ocicore.VolumeBackup)
	c.bootVolumes = make(map[string]ocicore.BootVolume)
	c.volumes = make(map[string]ocicore.Volume)
	return &c
}

// CreateVolumeBackup returns a fake response for CreateVolumeBackup
func (cc *BlockStorageClient) CreateVolumeBackup(ctx context.Context, request ocicore.CreateVolumeBackupRequest) (response ocicore.CreateVolumeBackupResponse, err error) {
	response = ocicore.CreateVolumeBackupResponse{}
	vol := ocicore.VolumeBackup{}
	ocid := string(uuid.NewUUID())
	vol.Id = &ocid
	cc.backupVolumes[ocid] = vol
	response.VolumeBackup = vol
	return response, nil
}

// GetVolumeBackup returns a fake response for GetVolumeBackup
func (cc *BlockStorageClient) GetVolumeBackup(ctx context.Context, request ocicore.GetVolumeBackupRequest) (response ocicore.GetVolumeBackupResponse, err error) {
	response = ocicore.GetVolumeBackupResponse{}
	vol := ocicore.VolumeBackup{
		LifecycleState: ocicore.VolumeBackupLifecycleStateAvailable,
	}
	response.VolumeBackup = vol
	return response, nil
}

// DeleteVolumeBackup returns a fake response for DeleteVolumeBackup
func (cc *BlockStorageClient) DeleteVolumeBackup(ctx context.Context, request ocicore.DeleteVolumeBackupRequest) (response ocicore.DeleteVolumeBackupResponse, err error) {
	response = ocicore.DeleteVolumeBackupResponse{}
	return response, nil
}

// UpdateVolumeBackup returns a fake response for UpdateVolumeBackup
func (cc *BlockStorageClient) UpdateVolumeBackup(ctx context.Context, request ocicore.UpdateVolumeBackupRequest) (response ocicore.UpdateVolumeBackupResponse, err error) {
	response = ocicore.UpdateVolumeBackupResponse{}
	return response, nil
}

// CreateVolume returns a fake response for CreateVolume
func (cc *BlockStorageClient) CreateVolume(ctx context.Context, request ocicore.CreateVolumeRequest) (response ocicore.CreateVolumeResponse, err error) {
	response = ocicore.CreateVolumeResponse{}
	vol := ocicore.Volume{}
	ocid := string(uuid.NewUUID())
	vol.Id = &ocid
	cc.volumes[ocid] = vol

	response.Volume = vol
	return response, nil
}

// GetBootVolume returns a fake response for GetBootVolume
func (cc *BlockStorageClient) GetBootVolume(ctx context.Context, request ocicore.GetBootVolumeRequest) (response ocicore.GetBootVolumeResponse, err error) {
	response = ocicore.GetBootVolumeResponse{}

	bv := ocicore.BootVolume{
		LifecycleState: ocicore.BootVolumeLifecycleStateAvailable,
	}
	response.BootVolume = bv
	return response, nil
}

// GetVolume returns a fake response for GetVolume
func (cc *BlockStorageClient) GetVolume(ctx context.Context, request ocicore.GetVolumeRequest) (response ocicore.GetVolumeResponse, err error) {
	response = ocicore.GetVolumeResponse{}
	ocid := string(uuid.NewUUID())
	vol := ocicore.Volume{
		LifecycleState: ocicore.VolumeLifecycleStateAvailable,
		Id:             &ocid,
	}
	response.Volume = vol
	return response, nil
}

// DeleteVolume returns a fake response for DeleteVolume
func (cc *BlockStorageClient) DeleteVolume(ctx context.Context, request ocicore.DeleteVolumeRequest) (response ocicore.DeleteVolumeResponse, err error) {
	response = ocicore.DeleteVolumeResponse{}
	return response, nil
}

// UpdateVolume returns a fake response for UpdateVolume
func (cc *BlockStorageClient) UpdateVolume(ctx context.Context, request ocicore.UpdateVolumeRequest) (response ocicore.UpdateVolumeResponse, err error) {
	response = ocicore.UpdateVolumeResponse{}
	return response, nil
}
