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
	"strconv"

	ocilb "github.com/oracle/oci-go-sdk/loadbalancer"

	"k8s.io/apimachinery/pkg/util/uuid"

	resourcescommon "github.com/oracle/oci-manager/pkg/controller/oci/resources/common"
)

const (
	// FakeWorkRequestID used as work request id for fake responses
	FakeWorkRequestID = "wrID123"
)

var (
	// fakeLoadBalancerID used as load balancer id for fake responses
	fakeLoadBalancerID = "lbID123"
)

// LoadBalancerClient implements common LoadBalancerClient to fake oci methods for unit tests.
type LoadBalancerClient struct {
	resourcescommon.LoadBalancerClientInterface

	backends      map[string]ocilb.Backend
	backendSets   map[string]ocilb.BackendSet
	certificates  map[string]ocilb.Certificate
	listeners     map[string]ocilb.Listener
	loadBalancers map[string]ocilb.LoadBalancer
	workRequests  map[string]ocilb.WorkRequest
}

// NewLoadBalancerClient returns a client that will respond with the provided objects.
// It shouldn't be used as a replacement for a real client and is mostly useful in simple unit tests.
func NewLoadBalancerClient() (flbc *LoadBalancerClient) {
	c := LoadBalancerClient{}
	c.backends = make(map[string]ocilb.Backend)
	c.backendSets = make(map[string]ocilb.BackendSet)
	c.certificates = make(map[string]ocilb.Certificate)
	c.listeners = make(map[string]ocilb.Listener)
	c.loadBalancers = make(map[string]ocilb.LoadBalancer)
	c.workRequests = make(map[string]ocilb.WorkRequest)
	return &c
}

// CreateBackend returns a fake response for CreateBackend
func (lbc *LoadBalancerClient) CreateBackend(ctx context.Context, request ocilb.CreateBackendRequest) (response ocilb.CreateBackendResponse, err error) {
	response = ocilb.CreateBackendResponse{}
	backend := ocilb.Backend{}
	backend.IpAddress = request.IpAddress
	name := *request.IpAddress + ":" + strconv.Itoa(*request.Port)
	backend.Name = &name
	backend.Port = request.Port
	backend.Backup = request.Backup
	backend.Drain = request.Drain
	backend.Offline = request.Offline
	backend.Weight = request.Weight

	wrID := FakeWorkRequestID
	response.OpcWorkRequestId = &wrID

	id := *request.LoadBalancerId + *request.BackendSetName + name
	lbc.backends[id] = backend

	return response, nil
}

// CreateBackendSet returns a fake response for CreateBackendSet
func (lbc *LoadBalancerClient) CreateBackendSet(ctx context.Context, request ocilb.CreateBackendSetRequest) (response ocilb.CreateBackendSetResponse, err error) {
	response = ocilb.CreateBackendSetResponse{}
	wrID := FakeWorkRequestID
	response.OpcWorkRequestId = &wrID

	backendSet := ocilb.BackendSet{}
	backendSet.Name = request.Name
	hc := ocilb.HealthChecker{}
	hc.Port = request.HealthChecker.Port
	hc.UrlPath = request.HealthChecker.UrlPath
	hc.Protocol = request.HealthChecker.Protocol

	backendSet.HealthChecker = &hc

	id := *request.LoadBalancerId + *request.Name
	lbc.backendSets[id] = backendSet

	return response, nil
}

// CreateCertificate returns a fake response for CreateCertificate
func (lbc *LoadBalancerClient) CreateCertificate(ctx context.Context, request ocilb.CreateCertificateRequest) (response ocilb.CreateCertificateResponse, err error) {
	response = ocilb.CreateCertificateResponse{}
	wrID := FakeWorkRequestID
	response.OpcWorkRequestId = &wrID

	cert := ocilb.Certificate{}
	cert.CertificateName = request.CertificateName
	cert.PublicCertificate = request.PublicCertificate
	cert.CaCertificate = request.CaCertificate

	id := *request.LoadBalancerId + *request.CertificateName
	lbc.certificates[id] = cert

	return response, nil
}

// CreateListener returns a fake response for CreateListener
func (lbc *LoadBalancerClient) CreateListener(ctx context.Context, request ocilb.CreateListenerRequest) (response ocilb.CreateListenerResponse, err error) {
	response = ocilb.CreateListenerResponse{}
	wrID := FakeWorkRequestID
	response.OpcWorkRequestId = &wrID

	listener := ocilb.Listener{}
	listener.Protocol = request.Protocol
	listener.Port = request.Port
	listener.Name = request.Name
	listener.DefaultBackendSetName = request.DefaultBackendSetName

	id := *request.LoadBalancerId + *request.Name
	lbc.listeners[id] = listener

	return response, nil
}

// CreateLoadBalancer returns a fake response for CreateLoadBalancer
func (lbc *LoadBalancerClient) CreateLoadBalancer(ctx context.Context, request ocilb.CreateLoadBalancerRequest) (response ocilb.CreateLoadBalancerResponse, err error) {
	response = ocilb.CreateLoadBalancerResponse{}
	wrID := FakeWorkRequestID
	response.OpcWorkRequestId = &wrID

	lb := ocilb.LoadBalancer{}
	lbID := string(uuid.NewUUID())
	lb.Id = &lbID

	return response, nil
}

// DeleteBackend returns a fake response for DeleteBackend
func (lbc *LoadBalancerClient) DeleteBackend(ctx context.Context, request ocilb.DeleteBackendRequest) (response ocilb.DeleteBackendResponse, err error) {
	lbID := *request.LoadBalancerId
	backendSetName := *request.BackendSetName
	backendName := *request.BackendName

	id := lbID + backendSetName + backendName
	delete(lbc.backends, id)

	response = ocilb.DeleteBackendResponse{}
	reqid := string(uuid.NewUUID())
	response.OpcRequestId = &reqid
	return response, nil
}

// DeleteBackendSet returns a fake response for DeleteBackendSet
func (lbc *LoadBalancerClient) DeleteBackendSet(ctx context.Context, request ocilb.DeleteBackendSetRequest) (response ocilb.DeleteBackendSetResponse, err error) {
	lbID := *request.LoadBalancerId
	backendSetName := *request.BackendSetName

	id := lbID + backendSetName
	delete(lbc.backendSets, id)

	response = ocilb.DeleteBackendSetResponse{}
	reqid := string(uuid.NewUUID())
	response.OpcRequestId = &reqid
	return response, nil
}

// DeleteCertificate returns a fake response for DeleteCertificate
func (lbc *LoadBalancerClient) DeleteCertificate(ctx context.Context, request ocilb.DeleteCertificateRequest) (response ocilb.DeleteCertificateResponse, err error) {
	lbID := *request.LoadBalancerId
	certName := *request.CertificateName

	id := lbID + certName
	delete(lbc.certificates, id)

	response = ocilb.DeleteCertificateResponse{}
	reqid := string(uuid.NewUUID())
	response.OpcRequestId = &reqid
	return response, nil
}

// DeleteListener returns a fake response for DeleteListener
func (lbc *LoadBalancerClient) DeleteListener(ctx context.Context, request ocilb.DeleteListenerRequest) (response ocilb.DeleteListenerResponse, err error) {
	lbID := *request.LoadBalancerId
	listenerName := *request.ListenerName

	id := lbID + listenerName
	delete(lbc.listeners, id)

	response = ocilb.DeleteListenerResponse{}
	reqid := string(uuid.NewUUID())
	response.OpcRequestId = &reqid
	return response, nil
}

// DeleteLoadBalancer returns a fake response for DeleteLoadBalancer
func (lbc *LoadBalancerClient) DeleteLoadBalancer(ctx context.Context, request ocilb.DeleteLoadBalancerRequest) (response ocilb.DeleteLoadBalancerResponse, err error) {
	lbID := *request.LoadBalancerId
	delete(lbc.loadBalancers, lbID)

	response = ocilb.DeleteLoadBalancerResponse{}
	reqid := string(uuid.NewUUID())
	response.OpcWorkRequestId = &reqid
	return response, nil
}

// GetBackend returns a fake response for GetBackend
func (lbc *LoadBalancerClient) GetBackend(ctx context.Context, request ocilb.GetBackendRequest) (response ocilb.GetBackendResponse, err error) {
	lbID := *request.LoadBalancerId
	backendSetName := *request.BackendSetName
	backendName := *request.BackendName

	id := lbID + backendSetName + backendName

	if be, ok := lbc.backends[id]; ok {
		response = ocilb.GetBackendResponse{}
		response.Backend = be
		return response, nil
	}
	response = ocilb.GetBackendResponse{}
	return response, servicefailure{Message: "Not found", Code: "NotAuthorizedOrNotFound"}
}

// GetBackendSet returns a fake response for GetBackendSet
func (lbc *LoadBalancerClient) GetBackendSet(ctx context.Context, request ocilb.GetBackendSetRequest) (response ocilb.GetBackendSetResponse, err error) {
	lbID := *request.LoadBalancerId
	backendSetName := *request.BackendSetName

	id := lbID + backendSetName

	if bs, ok := lbc.backendSets[id]; ok {
		response = ocilb.GetBackendSetResponse{}
		response.BackendSet = bs
		return response, nil
	}
	response = ocilb.GetBackendSetResponse{}
	return response, servicefailure{Message: "Not found", Code: "NotAuthorizedOrNotFound"}
}

// GetLoadBalancer returns a fake response for GetLoadBalancer
func (lbc *LoadBalancerClient) GetLoadBalancer(ctx context.Context, request ocilb.GetLoadBalancerRequest) (response ocilb.GetLoadBalancerResponse, err error) {
	lb := ocilb.LoadBalancer{
		Id: &fakeLoadBalancerID,
	}
	response = ocilb.GetLoadBalancerResponse{}
	response.LoadBalancer = lb
	return response, nil
}

// GetWorkRequest returns a fake response for GetWorkRequest
func (lbc *LoadBalancerClient) GetWorkRequest(ctx context.Context, request ocilb.GetWorkRequestRequest) (response ocilb.GetWorkRequestResponse, err error) {
	wr := ocilb.WorkRequest{
		LifecycleState: ocilb.WorkRequestLifecycleStateSucceeded,
		LoadBalancerId: &fakeLoadBalancerID,
	}
	response = ocilb.GetWorkRequestResponse{
		WorkRequest: wr,
	}
	return response, nil
}

// UpdateBackend returns a fake response for UpdateBackend
func (lbc *LoadBalancerClient) UpdateBackend(ctx context.Context, request ocilb.UpdateBackendRequest) (response ocilb.UpdateBackendResponse, err error) {
	response = ocilb.UpdateBackendResponse{}
	workRequestID := FakeWorkRequestID
	response.OpcWorkRequestId = &workRequestID
	return response, nil
}

// UpdateBackendSet returns a fake response for UpdateBackendSet
func (lbc *LoadBalancerClient) UpdateBackendSet(ctx context.Context, request ocilb.UpdateBackendSetRequest) (response ocilb.UpdateBackendSetResponse, err error) {
	response = ocilb.UpdateBackendSetResponse{}
	workRequestID := FakeWorkRequestID
	response.OpcWorkRequestId = &workRequestID
	return response, nil
}

// UpdateListener returns a fake response for UpdateListener
func (lbc *LoadBalancerClient) UpdateListener(ctx context.Context, request ocilb.UpdateListenerRequest) (response ocilb.UpdateListenerResponse, err error) {
	lbID := *request.LoadBalancerId
	backendSetName := *request.DefaultBackendSetName

	id := lbID + backendSetName

	if _, ok := lbc.listeners[id]; ok {
		response = ocilb.UpdateListenerResponse{}
		workRequestID := FakeWorkRequestID
		response.OpcWorkRequestId = &workRequestID
		return response, nil
	}
	response = ocilb.UpdateListenerResponse{}
	return response, servicefailure{Message: "Not found", Code: "NotAuthorizedOrNotFound"}
}

// UpdateLoadBalancer returns a fake response for UpdateLoadBalancer
func (lbc *LoadBalancerClient) UpdateLoadBalancer(ctx context.Context, request ocilb.UpdateLoadBalancerRequest) (response ocilb.UpdateLoadBalancerResponse, err error) {
	lbID := *request.LoadBalancerId

	if _, ok := lbc.loadBalancers[lbID]; ok {
		response = ocilb.UpdateLoadBalancerResponse{}
		workRequestID := FakeWorkRequestID
		response.OpcWorkRequestId = &workRequestID
		return response, nil
	}
	response = ocilb.UpdateLoadBalancerResponse{}
	return response, servicefailure{Message: "Not found", Code: "NotAuthorizedOrNotFound"}
}
