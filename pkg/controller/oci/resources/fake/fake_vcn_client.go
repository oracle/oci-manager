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
	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
	"k8s.io/apimachinery/pkg/util/uuid"
)

// VcnClient implements common VcnClientInterface to fake oci methods for unit tests.
type VcnClient struct {
	resourcescommon.VcnClientInterface

	dhcpOptions      map[string]ocicore.DhcpOptions
	internetGateways map[string]ocicore.InternetGateway
	routeTables      map[string]ocicore.RouteTable
	securityLists    map[string]ocicore.SecurityList
	subnets          map[string]ocicore.Subnet
	vcns             map[string]ocicore.Vcn
}

// NewVcnClient returns a client that will respond with the provided objects.
// It shouldn't be used as a replacement for a real client and is mostly useful in simple unit tests.
func NewVcnClient() (fvcnc *VcnClient) {
	c := VcnClient{}

	c.dhcpOptions = make(map[string]ocicore.DhcpOptions)
	c.internetGateways = make(map[string]ocicore.InternetGateway)
	c.routeTables = make(map[string]ocicore.RouteTable)
	c.securityLists = make(map[string]ocicore.SecurityList)
	c.subnets = make(map[string]ocicore.Subnet)
	c.vcns = make(map[string]ocicore.Vcn)

	return &c
}

// CreateDhcpOptions returns a fake response for CreateDhcpOptions
func (vcnc *VcnClient) CreateDhcpOptions(ctx context.Context, request ocicore.CreateDhcpOptionsRequest) (response ocicore.CreateDhcpOptionsResponse, err error) {

	response = ocicore.CreateDhcpOptionsResponse{}
	vcn := ocicore.Vcn{}
	ocid := string(uuid.NewUUID())
	vcn.Id = &ocid
	vcnc.vcns[ocid] = vcn

	response.Id = &ocid
	return response, nil
}

// DeleteDhcpOptions returns a fake response for DeleteDhcpOptions
func (vcnc *VcnClient) DeleteDhcpOptions(ctx context.Context, request ocicore.DeleteDhcpOptionsRequest) (response ocicore.DeleteDhcpOptionsResponse, err error) {
	//id := *request.DhcpId
	//delete(vcnc.dhcpOptions, id)
	response = ocicore.DeleteDhcpOptionsResponse{}
	return response, nil

}

// GetDhcpOptions returns a fake response for CreateDhcpOptions
func (vcnc *VcnClient) GetDhcpOptions(ctx context.Context, request ocicore.GetDhcpOptionsRequest) (response ocicore.GetDhcpOptionsResponse, err error) {
	response = ocicore.GetDhcpOptionsResponse{}
	return response, nil
}

// UpdateDhcpOptions returns a fake response for UpdateDhcpOptions
func (vcnc *VcnClient) UpdateDhcpOptions(ctx context.Context, request ocicore.UpdateDhcpOptionsRequest) (response ocicore.UpdateDhcpOptionsResponse, err error) {
	response = ocicore.UpdateDhcpOptionsResponse{}
	return response, nil
}

// CreateInternetGateway returns a fake response for CreateInternetGateway
func (vcnc *VcnClient) CreateInternetGateway(ctx context.Context, request ocicore.CreateInternetGatewayRequest) (response ocicore.CreateInternetGatewayResponse, err error) {
	response = ocicore.CreateInternetGatewayResponse{}

	ig := ocicore.InternetGateway{}
	ocid := string(uuid.NewUUID())
	ig.Id = &ocid
	vcnc.internetGateways[ocid] = ig

	response.InternetGateway = ig
	return response, nil
}

// UpdateInternetGateway returns a fake response for UpdateInternetGateway
func (vcnc *VcnClient) UpdateInternetGateway(ctx context.Context, request ocicore.UpdateInternetGatewayRequest) (response ocicore.UpdateInternetGatewayResponse, err error) {
	response = ocicore.UpdateInternetGatewayResponse{}

	if ig, ok := vcnc.internetGateways[*request.IgId]; ok {
		response.InternetGateway = ig
		return response, nil
	}
	return response, servicefailure{Message: "Not found", Code: "NotAuthorizedOrNotFound"}
}

// DeleteInternetGateway returns a fake response for DeleteInternetGateway
func (vcnc *VcnClient) DeleteInternetGateway(ctx context.Context, request ocicore.DeleteInternetGatewayRequest) (response ocicore.DeleteInternetGatewayResponse, err error) {

	id := *request.IgId
	delete(vcnc.internetGateways, id)

	response = ocicore.DeleteInternetGatewayResponse{}
	return response, nil
}

// GetInternetGateway returns a fake response for GetInternetGateway
func (vcnc *VcnClient) GetInternetGateway(ctx context.Context, request ocicore.GetInternetGatewayRequest) (response ocicore.GetInternetGatewayResponse, err error) {
	response = ocicore.GetInternetGatewayResponse{}

	if ig, ok := vcnc.internetGateways[*request.IgId]; ok {
		response.InternetGateway = ig
		return response, nil
	}
	return response, servicefailure{Message: "Not found", Code: "NotAuthorizedOrNotFound"}
}

// CreateSubnet returns a fake response for CreateSubnet
func (vcnc *VcnClient) CreateSubnet(ctx context.Context, request ocicore.CreateSubnetRequest) (response ocicore.CreateSubnetResponse, err error) {

	subnet := ocicore.Subnet{}
	ocid := string(uuid.NewUUID())
	subnet.Id = &ocid
	vcnc.subnets[ocid] = subnet

	response = ocicore.CreateSubnetResponse{}
	response.Subnet = subnet
	return response, nil
}

// DeleteSubnet returns a fake response for DeleteSubnet
func (vcnc *VcnClient) DeleteSubnet(ctx context.Context, request ocicore.DeleteSubnetRequest) (response ocicore.DeleteSubnetResponse, err error) {

	id := *request.SubnetId
	delete(vcnc.subnets, id)
	response = ocicore.DeleteSubnetResponse{}
	return response, nil
}

// UpdateSubnet returns a fake response for UpdateSubnet
func (vcnc *VcnClient) UpdateSubnet(ctx context.Context, request ocicore.UpdateSubnetRequest) (response ocicore.UpdateSubnetResponse, err error) {

	if subnet, ok := vcnc.subnets[*request.SubnetId]; ok {
		response = ocicore.UpdateSubnetResponse{}
		response.Subnet = subnet
		return response, nil
	}
	response = ocicore.UpdateSubnetResponse{}
	return response, servicefailure{Message: "Not found", Code: "NotAuthorizedOrNotFound"}
}

// GetSubnet returns a fake response for GetSubnet
func (vcnc *VcnClient) GetSubnet(ctx context.Context, request ocicore.GetSubnetRequest) (response ocicore.GetSubnetResponse, err error) {
	response = ocicore.GetSubnetResponse{}

	if subnet, ok := vcnc.subnets[*request.SubnetId]; ok {
		response.Subnet = subnet
		return response, nil
	}
	return response, servicefailure{Message: "Not found", Code: "NotAuthorizedOrNotFound"}
}

// CreateSecurityList returns a fake response for CreateSecurityList
func (vcnc *VcnClient) CreateSecurityList(ctx context.Context, request ocicore.CreateSecurityListRequest) (response ocicore.CreateSecurityListResponse, err error) {

	sl := ocicore.SecurityList{}
	ocid := string(uuid.NewUUID())
	sl.Id = &ocid
	vcnc.securityLists[ocid] = sl

	response = ocicore.CreateSecurityListResponse{}
	response.SecurityList = sl
	return response, nil
}

// UpdateSecurityList returns a fake response for UpdateSecurityList
func (vcnc *VcnClient) UpdateSecurityList(ctx context.Context, request ocicore.UpdateSecurityListRequest) (response ocicore.UpdateSecurityListResponse, err error) {

	if sl, ok := vcnc.securityLists[*request.SecurityListId]; ok {
		response = ocicore.UpdateSecurityListResponse{}
		response.SecurityList = sl
		return response, nil
	}
	response = ocicore.UpdateSecurityListResponse{}
	return response, servicefailure{Message: "Not found", Code: "NotAuthorizedOrNotFound"}
}

// DeleteSecurityList returns a fake response for DeleteSecurityList
func (vcnc *VcnClient) DeleteSecurityList(ctx context.Context, request ocicore.DeleteSecurityListRequest) (response ocicore.DeleteSecurityListResponse, err error) {
	response = ocicore.DeleteSecurityListResponse{}
	return response, nil
}

// GetSecurityList returns a fake response for GetSecurityList
func (vcnc *VcnClient) GetSecurityList(ctx context.Context, request ocicore.GetSecurityListRequest) (response ocicore.GetSecurityListResponse, err error) {
	return ocicore.GetSecurityListResponse{}, nil
}

// CreateRouteTable returns a fake response for CreateRouteTable
func (vcnc *VcnClient) CreateRouteTable(ctx context.Context, request ocicore.CreateRouteTableRequest) (response ocicore.CreateRouteTableResponse, err error) {

	rt := ocicore.RouteTable{}
	ocid := string(uuid.NewUUID())
	rt.Id = &ocid
	vcnc.routeTables[ocid] = rt

	response = ocicore.CreateRouteTableResponse{}
	response.RouteTable = rt
	return response, nil
}

// UpdateRouteTable returns a fake response for UpdateRouteTable
func (vcnc *VcnClient) UpdateRouteTable(ctx context.Context, request ocicore.UpdateRouteTableRequest) (response ocicore.UpdateRouteTableResponse, err error) {

	if rt, ok := vcnc.routeTables[*request.RtId]; ok {
		response = ocicore.UpdateRouteTableResponse{}
		response.RouteTable = rt
		return response, nil
	}
	response = ocicore.UpdateRouteTableResponse{}
	return response, servicefailure{Message: "Not found", Code: "NotAuthorizedOrNotFound"}
}

// DeleteRouteTable returns a fake response for DeleteRouteTable
func (vcnc *VcnClient) DeleteRouteTable(ctx context.Context, request ocicore.DeleteRouteTableRequest) (response ocicore.DeleteRouteTableResponse, err error) {

	id := *request.RtId
	delete(vcnc.routeTables, id)
	response = ocicore.DeleteRouteTableResponse{}
	return response, nil
}

// GetRouteTable returns a fake response for GetRouteTable
func (vcnc *VcnClient) GetRouteTable(ctx context.Context, request ocicore.GetRouteTableRequest) (response ocicore.GetRouteTableResponse, err error) {
	response = ocicore.GetRouteTableResponse{}
	if rt, ok := vcnc.routeTables[*request.RtId]; ok {
		response.RouteTable = rt
		return response, nil
	}
	return response, servicefailure{Message: "Not found", Code: "NotAuthorizedOrNotFound"}
}

// CreateVcn returns a fake response for CreateVcn
func (vcnc *VcnClient) CreateVcn(ctx context.Context, request ocicore.CreateVcnRequest) (response ocicore.CreateVcnResponse, err error) {

	response = ocicore.CreateVcnResponse{}
	vcn := ocicore.Vcn{}
	ocid := string(uuid.NewUUID())
	vcn.Id = &ocid
	vcn.DisplayName = request.DisplayName
	vcn.CidrBlock = request.CidrBlock
	vcn.DnsLabel = request.DnsLabel

	vcnc.vcns[ocid] = vcn

	response.Vcn = vcn
	return response, nil
}

// DeleteVcn returns a fake response for DeleteVcn
func (vcnc *VcnClient) DeleteVcn(ctx context.Context, request ocicore.DeleteVcnRequest) (response ocicore.DeleteVcnResponse, err error) {

	id := *request.VcnId
	delete(vcnc.vcns, id)
	response = ocicore.DeleteVcnResponse{}
	return response, nil

}

// GetVcn returns a fake response for GetVcn
func (vcnc *VcnClient) GetVcn(ctx context.Context, request ocicore.GetVcnRequest) (response ocicore.GetVcnResponse, err error) {
	response = ocicore.GetVcnResponse{}
	if vcn, ok := vcnc.vcns[*request.VcnId]; ok {
		response.Vcn = vcn
		return response, nil
	}
	return response, servicefailure{Message: "Not found", Code: "NotAuthorizedOrNotFound"}
}

// UpdateVcn returns a fake response for UpdateVcn
func (vcnc *VcnClient) UpdateVcn(ctx context.Context, request ocicore.UpdateVcnRequest) (response ocicore.UpdateVcnResponse, err error) {
	if vcn, ok := vcnc.vcns[*request.VcnId]; ok {
		response = ocicore.UpdateVcnResponse{}
		response.Vcn = vcn
		return response, nil
	}
	response = ocicore.UpdateVcnResponse{}
	return response, servicefailure{Message: "Not found", Code: "NotAuthorizedOrNotFound"}
}

// GetVnic returns a fake response for GetVnic
func (vcnc *VcnClient) GetVnic(ctx context.Context, request ocicore.GetVnicRequest) (response ocicore.GetVnicResponse, err error) {
	response = ocicore.GetVnicResponse{}
	isPrimary := true
	vnic := ocicore.Vnic{
		IsPrimary: &isPrimary,
	}
	response.Vnic = vnic
	return response, nil
}
