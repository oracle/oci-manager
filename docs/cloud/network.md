# Network

## Specification
[network_types.go](../../pkg/apis/cloud.k8s.io/v1alpha1/network_types.go)

## Implementation
- VCN - specify using cidrBlock attribute or defaults to 10.0.0.0/16
- Route table - defaults to route everything thru internet gateway

  future: add route rules to peering gateways or dynamic routing gateways
- Internet gateway: enabled
