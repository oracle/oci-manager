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

	"k8s.io/apimachinery/pkg/util/uuid"

	ociid "github.com/oracle/oci-go-sdk/identity"
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
)

// IdentityClient implements common IdentityClient to fake oci methods for unit tests.
type IdentityClient struct {
	resourcescommon.IdentityClientInterface

	policies     map[string]ociid.Policy
	compartments map[string]ociid.Compartment
}

// NewIdentityClient returns a client that will respond with the provided objects.
// It shouldn't be used as a replacement for a real client and is mostly useful in simple unit tests.
func NewIdentityClient() (fcc *IdentityClient) {
	c := IdentityClient{}

	c.policies = make(map[string]ociid.Policy)
	c.compartments = make(map[string]ociid.Compartment)
	return &c
}

// GetCompartment returns a fake response for GetCompartment
func (cc *IdentityClient) GetCompartment(ctx context.Context, request ociid.GetCompartmentRequest) (response ociid.GetCompartmentResponse, err error) {
	name := "compartment.test1"
	id := "fakeCompartmentOCID"
	compartment := ociid.Compartment{
		Name: &name,
		Id:   &id,
	}
	response = ociid.GetCompartmentResponse{}
	response.Compartment = compartment
	return response, nil
}

// ListCompartments returns a fake response for ListCompartments
func (cc *IdentityClient) ListCompartments(ctx context.Context, request ociid.ListCompartmentsRequest) (response ociid.ListCompartmentsResponse, err error) {
	response = ociid.ListCompartmentsResponse{}
	testName := "compartment.test1"
	testID := "testociid.compartment.test1"
	rootCompartment := "compartment.root.test1"
	testDesc := "test"
	item := ociid.Compartment{
		Name:          &testName,
		CompartmentId: &rootCompartment,
		Id:            &testID,
		Description:   &testDesc,
	}
	response.Items = append(response.Items, item)
	return response, nil
}

// ListAvailabilityDomains returns a fake response for ListAvailabilityDomains
func (cc *IdentityClient) ListAvailabilityDomains(ctx context.Context, request ociid.ListAvailabilityDomainsRequest) (response ociid.ListAvailabilityDomainsResponse, err error) {
	response = ociid.ListAvailabilityDomainsResponse{}
	testName := "ad.test1"
	rootCompartment := "compartment.root.test1"
	item := ociid.AvailabilityDomain{
		Name:          &testName,
		CompartmentId: &rootCompartment,
	}
	response.Items = append(response.Items, item)
	return response, nil
}

// CreatePolicy returns a fake response for CreatePolicy
func (cc *IdentityClient) CreatePolicy(ctx context.Context, request ociid.CreatePolicyRequest) (response ociid.CreatePolicyResponse, err error) {

	response = ociid.CreatePolicyResponse{}
	policy := ociid.Policy{}
	ocid := string(uuid.NewUUID())
	policy.Id = &ocid
	cc.policies[ocid] = policy

	response.Id = &ocid
	return response, nil
}

// DeletePolicy returns a fake response for DeletePolicy
func (cc *IdentityClient) DeletePolicy(ctx context.Context, request ociid.DeletePolicyRequest) (response ociid.DeletePolicyResponse, err error) {
	//id := *request.DhcpId
	//delete(vcnc.dhcpOptions, id)
	response = ociid.DeletePolicyResponse{}
	return response, nil

}

// GetPolicy returns a fake response for GetPolicy
func (cc *IdentityClient) GetPolicy(ctx context.Context, request ociid.GetPolicyRequest) (response ociid.GetPolicyResponse, err error) {
	response = ociid.GetPolicyResponse{}
	return response, nil
}

// UpdatePolicy returns a fake response for UpdatePolicy
func (cc *IdentityClient) UpdatePolicy(ctx context.Context, request ociid.UpdatePolicyRequest) (response ociid.UpdatePolicyResponse, err error) {
	response = ociid.UpdatePolicyResponse{}
	return response, nil
}
