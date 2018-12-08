# LoadBalancer

## Specification
[loadbalancer_types.go](../../pkg/apis/cloud.k8s.io/v1alpha1/loadbalancer_types.go)

## Implementation
- Creates Subnets using a security selector
- Creates Backends using a compute selector
  - multiple computes can be selected for different versions of your app
  - labelWeightMap attribute to control traffic to different computes / versions
- Listeners allows multiple front-end / client-facing ports
  with ssl certificate struct to model certs, private key, passphrase
- balance mode: round-robin or least-connections
- bandwidthMbps: used for size/speed/cost of load-balancer - defaults to 100Mbps
