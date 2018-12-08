# Security

## Specification
[security_types.go](../../pkg/apis/cloud.k8s.io/v1alpha1/security_types.go)

## Implementation
- Maintains Security Lists in VCN matching network selector
- Compute and loadbalancer have security selector to set when creating their subnets
- Standard allow rules for ingress and egress.
- Defaults to allow only 22 and all egress
