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
	ocidb "github.com/oracle/oci-go-sdk/database"

	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
)

var (
	fakeDbID = "123"
	fakeName = "fakeDB"
	one      = 1
)

// DatabaseClient implements common DatabaseClient to fake oci methods for unit tests.
type DatabaseClient struct {
	resourcescommon.DatabaseClientInterface

	autonomousdatabases map[string]ocidb.AutonomousDatabase
}

// NewDatabaseClient returns a client that will respond with the provided objects.
// It shouldn't be used as a replacement for a real client and is mostly useful in simple unit tests.
func NewDatabaseClient() (dbc *DatabaseClient) {
	c := DatabaseClient{}
	c.autonomousdatabases = make(map[string]ocidb.AutonomousDatabase)
	return &c
}

// CreateAutonomousDatabase returns a fake response
func (dbc *DatabaseClient) CreateAutonomousDatabase(ctx context.Context, request ocidb.CreateAutonomousDatabaseRequest) (response ocidb.CreateAutonomousDatabaseResponse, err error) {

	response = ocidb.CreateAutonomousDatabaseResponse{}

	db := ocidb.AutonomousDatabase{}
	db.DbName = request.DbName

	dbc.autonomousdatabases[fakeDbID] = db

	return response, nil
}

func (dbc *DatabaseClient) DeleteAutonomousDatabase(ctx context.Context, request ocidb.DeleteAutonomousDatabaseRequest) (response ocidb.DeleteAutonomousDatabaseResponse, err error) {

	response = ocidb.DeleteAutonomousDatabaseResponse{}

	return response, nil
}

func (dbc *DatabaseClient) GetAutonomousDatabase(ctx context.Context, request ocidb.GetAutonomousDatabaseRequest) (response ocidb.GetAutonomousDatabaseResponse, err error) {

	response = ocidb.GetAutonomousDatabaseResponse{
		AutonomousDatabase: ocidb.AutonomousDatabase{
			Id:                   &fakeDbID,
			DataStorageSizeInTBs: &one,
			DbName:               &fakeName,
			CpuCoreCount:         &one,
		},
	}

	return response, nil
}

func (dbc *DatabaseClient) ListAutonomousDatabases(ctx context.Context, request ocidb.ListAutonomousDatabasesRequest) (response ocidb.ListAutonomousDatabasesResponse, err error) {

	response = ocidb.ListAutonomousDatabasesResponse{}

	return response, nil
}

func (dbc *DatabaseClient) StartAutonomousDatabase(ctx context.Context, request ocidb.StartAutonomousDatabaseRequest) (response ocidb.StartAutonomousDatabaseResponse, err error) {

	response = ocidb.StartAutonomousDatabaseResponse{}

	return response, nil
}

func (dbc *DatabaseClient) StopAutonomousDatabase(ctx context.Context, request ocidb.StopAutonomousDatabaseRequest) (response ocidb.StopAutonomousDatabaseResponse, err error) {

	response = ocidb.StopAutonomousDatabaseResponse{}

	return response, nil
}

func (dbc *DatabaseClient) UpdateAutonomousDatabase(ctx context.Context, request ocidb.UpdateAutonomousDatabaseRequest) (response ocidb.UpdateAutonomousDatabaseResponse, err error) {

	response = ocidb.UpdateAutonomousDatabaseResponse{}

	return response, nil
}
