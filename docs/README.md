# OCI Manager Documentation

## Introduction
OCIM leverages the extensibility features of Kubernetes to apply cloud-native principles and best practices for orchestrating OCI infrastructure resources. It declares a CRD ([Custom Resource Definition][1]) for each supported OCI resource type, along with corresponding controllers that are used for reconciling a desired state for those resources.


## Setup and Installation

To get OCIM running you will need to do the following:

 1. Prepare a Kubernetes cluster v1.10+
 2. Setup configuration and choose OCI authentication setup.
 3. Deploy OCIM application.

For detailed instructions see the [setup documentation](setup.md).



## Try it out

Although not required, it's recommended that you first associate a Kubernetes namespace with an OCI compartment of the same name. Assuming you have a compartment with a name `example` you can do so by running:

```bash
kubectl create namespace example
kubectl label namespace example oci-compartment=true
```

The namespace label will create a Compartment object in that namespace.  If the compartment did not exist in OCI, it will be created. If the compartment already existed, it will be used. To verify `kubectl get compartments -n example` and you should see 1 compartment resource with the name `example`.

The [examples/resources/v1alpha1](../examples/resources/v1alpha1) directory contains an example for each OCI resource type currently supported in OCIM. Review and edit an examples, then execute `kubectl apply -f <filename>` to create a resource. To delete the resource execute `kubectl delete -f <filename>`. For example, try the vcn.yaml.


## Use-cases
OCIM can be considered as a building block for these higher level use-cases:

- [Cloud Orchestration](cloud/README.md)

[1]: https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/
