# Cloud Orchestration

## Overview

OCI Manager (OCIM) defines a cloud API specification using Kubernetes CRDs and then implements it using controllers and leverages OCI resources. It follows a Kubernetes operator-style controller pattern where the cloud custom resources and controllers own and operate various OCI spec resources. The purpose of this layer is to provide:

- **Simplification** when dealing with lower level infrastructure resources by providing re-usable patterns of grouped resources as a single simplified resource.  
- **Abstraction** allowing for other implementations of the cloud API spec besides OCIM and providing workload portability.


## Cloud API Specification

- [Network](network.md) is a generic specification of an infrastructure network that includes CIDR ranges, routing, internet gateway connectivity etc.
- [Security](security.md)
captures ingress and egress rules that can be applied to multiple Compute and LoadBalancer objects using label selectors.
- [Compute](compute.md)
is a set of cloud instances defined using a single template. Uses replicas to scale up/down, similar to the concept of a ReplicaSet for containers.
- [LoadBalancer](loadbalancer.md)
is load balancer of cloud instances, similar to Service objects for containers.


## Getting Started

Start by reviewing and trying out a [basic stack example](../../examples/cloud/basic-stack.yaml) composed of cloud resources.
