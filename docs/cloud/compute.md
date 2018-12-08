# Compute

## Specification
[compute_types.go](../../pkg/apis/cloud.k8s.io/v1alpha1/compute_types.go)

## Implementation
- Creates Subnets using a security selector
- Creates Instances using replicas and template
- Template struct:
  - OS type and version to match an image
  - Resource Requirements (reused from kubernetes)
  to scope minimum (requests) and maximum (limits) for cpu, memory and
  storage - both network-attached and ephemeral -
  to determine shape / size / type of compute
  - SSH-keys: array of ssh public keys put on the instance to allow remote access
  - User-data: cloud-init executed script to install your app / artifact
- explicit shape and image via annotations:
  - oci.oracle.com/image=abc
  - oci.oracle.com/shape=abc
- Volumes: array of:
  - type: network-attached or ephemeral
  - size: amount to partition from available
  - mount point: path where the volume gets mounted
  - fstab options: for the fstab entry
