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
package v1alpha1

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	common "github.com/oracle/oci-manager/pkg/apis/ocicommon.oracle.com/v1alpha1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

const (
	LoadBalancerKind           = "LoadBalancer"
	LoadBalancerResourcePlural = "loadbalancers"
	LoadBalancerControllerName = "loadbalancers"
)

var minTCPPort = float64(1)
var maxTCPPort = float64(65535)

// LoadBalancerValidation describes the loadbalancer validation schema
var LoadBalancerValidation = apiextv1beta1.CustomResourceValidation{
	OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
		Properties: map[string]apiextv1beta1.JSONSchemaProps{
			"metadata": common.MetaDataValidation,
			"spec": {
				Required: []string{"listeners", "backendPort", "computeSelector", "securitySelector"},
				Properties: map[string]apiextv1beta1.JSONSchemaProps{
					"backendPort": {
						Type:    common.ValidationTypeInteger,
						Minimum: &minTCPPort,
						Maximum: &maxTCPPort,
					},
					"balanceMode": {
						Type:    common.ValidationTypeString,
						Pattern: common.AnyStringValidationRegex,
					},
					"listeners": {
						Type: common.ValidationTypeArray,
					},
				},
			},
		},
	},
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//LoadBalancer describes LoadBalancer
type LoadBalancer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              LoadBalancerSpec   `json:"spec"`
	Status            LoadBalancerStatus `json:"status,omitempty"`
}

type LoadBalancerSpec struct {
	// tcp port which backend servers use to serve traffic
	BackendPort int `json:"backendPort"`

	// balance mode defaults to ROUND_ROBIN
	BalanceMode string `json:"balanceMode,omitempty"`

	// network bandwidth maximum for loadbalancer
	BandwidthMbps string `json:"bandwidthMbps,omitempty"`

	// label selector of computes to backend instances
	ComputeSelector map[string]string `json:"computeSelector"`

	// health check parameters
	HealthCheck HealthCheck `json:"healthCheck,omitempty"`

	// configure to allow public/internet access
	IsPrivate bool `json:"isPrivate,omitempty"`

	// map of compute labels to weight
	LabelWeightMap map[string]int `json:"labelWeightMap"`

	// tcp ports which external clients will connect to
	Listeners []Listener `json:"listeners"`

	// selector to link security rule sets to subnets for the load balancer
	SecuritySelector map[string]string `json:"securitySelector,omitempty"`

	// name of cookie to use for session persistence
	SessionPersistenceCookie string `json:"sessionPersistenceCookie,omitempty"`

	Env       []apiv1.EnvVar             `json:"env,omitempty"`
	Resources apiv1.ResourceRequirements `json:"resources,omitempty"`
}

type Listener struct {
	Port     int    `json:"port,omitempty"`
	Protocol string `json:"protocol,omitempty"`

	// Optional
	IdleTimeoutSec int            `json:"idleTimeoutSec,omitempty"`
	SSLCertificate SSLCertificate `json:"sslCertificate,omitempty"`
}

// SSL Certificate for https traffic
type SSLCertificate struct {
	Certificate   string `json:"certificate,omitempty"`
	PrivateKey    string `json:"privateKey,omitempty"`
	Passphrase    string `json:"passphrase,omitempty"`
	CACertificate string `json:"caCertificate,omitempty"`
}

// HealthCheck describes health checker of the backend set
type HealthCheck struct {
	Protocol string `url:"-" header:"-" json:"protocol"`
	URLPath  string `url:"-" header:"-" json:"urlPath"`

	// Optional
	IntervalInMillis  int    `url:"-" header:"-" json:"intervalInMillis,omitempty"`  // Default: 10000
	Port              int    `url:"-" header:"-" json:"port,omitempty"`              // Default: 0
	ResponseBodyRegex string `url:"-" header:"-" json:"responseBodyRegex,omitempty"` // Default: ".*"
	Retries           int    `url:"-" header:"-" json:"retries,omitempty"`           // Default: 3
	ReturnCode        int    `url:"-" header:"-" json:"returnCode,omitempty"`        // Default: 200
	TimeoutInMillis   int    `url:"-" header:"-" json:"timeoutInMillis,omitempty"`   // Default: 3000
}

type LoadBalancerStatus struct {
	OperatorStatus

	IPAddress     string `json:"ipAddress"`
	DnsRecordName string `json:"dnsRecordName"`
	Network       string `json:"network"`

	// one-time randomized array of availability zones/domains for even distribution
	AvailabilityZones []string `json:"availabilityZones"`

	Instances          int `json:"instances"`
	AvailableInstances int `json:"availableInstances"`
	UnhealthyInstances int `json:"unhealthyInstances"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//LoadBalancerList is a list of LoadBalancer resources
type LoadBalancerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []LoadBalancer `json:"items"`
}
