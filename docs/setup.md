# OCI Manager Setup

## Before you begin

* To deploy OCIM you will need a Kubernetes cluster v1.11+ or higher. For information on how to install and run a Kubernetes cluster see [Kubernetes Setup](https://kubernetes.io/docs/setup/). It is highly recommended that you create a separate cluster for trying out OCIM. If you are using minikube, ensure you have the latest version and add `--kubernetes-version=v1.11.3` flag.
* Download and configure `kubectl` to access your kubernetes cluster and set the current context
* Install and configure [OCI CLI](https://docs.cloud.oracle.com/iaas/Content/API/SDKDocs/cliinstall.htm) (not required if the Kubernetes cluster is in OCI and you will use instance or service principles or if you want to setup the configuration and authentication manually)

## Deploy outside of Kubernetes cluster using Docker

The simplest and safest way to try OCI manager is to run it as a docker image outside of the Kubernetes cluster to avoid uploading OCI credentials in the Kubernetes cluster. To run OCI Manager:

First create the oci-system namespace:
```bash
$ kubectl create namespace oci-system
```

Ensure you have the latest oci-manager image in docker:
```bash
$ docker pull phx.ocir.io/k8sfed/oci-manager
```

Start oci-manager in a Docker container:
```bash
$ docker run --name oci-manager \
  -v $HOME/.oci:$HOME/.oci -e OCICONFIG=$HOME/.oci/config \
  -v $HOME/.kube:$HOME/.kube -e KUBECONFIG=$HOME/.kube/config  \
  -v $HOME/.minikube:$HOME/.minikube \
  phx.ocir.io/k8sfed/oci-manager
```

This command will mount and use your Kubernetes & OCI configuration and credentials from your local workstation configuration.

* To run it in the background add `-d` flag
* If you need to restart it again do `docker start oci-manager`
* If you need to upgrade it, delete the container first with `docker rm oci-manager`, then execute the `docker pull ...` and `docker run ...` commands from above again.
* If you are not using minikube remove the last volume mount, it's only needed for minikube since that's the directory where the Kubernetes client certificates reside


## Deploy inside OCI Kubernetes cluster using Instance Principals authentication

To create the namespace and RBAC run:
```bash
$ kubectl apply -f deploy/oci-manager-rbac.yaml
```

> NOTE that you will need cluster admin RBAC permissions for oci-manager

*Instance Principles* can be used when using a Kubernetes cluster running inside an OCI environment.

Deploy the OCIM application when using Instance Principles authentication:
```bash
$ kubectl apply -f deploy/oci-manager-ipr.yaml
```

> TODO instructions for creating an OCI policy


## Deploy inside any Kubernetes cluster using OCI API key authentication

> WARNING! Storing OCI API key inside Kubernetes by default is not secure and anyone with access to the Kubernetes cluster and read permissions for the secret will be able to retrieve the OCI API key

To create the namespace and RBAC run:
```bash
$ kubectl apply -f deploy/oci-manager-rbac.yaml
```

Configuration and authentication setup is needed since OCIM requires access to OCI API. *API Key* can be used when using any other Kubernetes cluster NOT running in OCI environment. To leverage your existing OCI CLI setup and generate the Kubernetes Secret and ConfigMap based on your home .oci/config file run:

```bash
$ deploy/setup.sh
```

This script will create a Secret with your OCI api key and a ConfigMap with your oci config file inside the `oci-system` namespace.

Deploy the OCIM application when using API key authentication:
```bash
$ kubectl apply -f deploy/oci-manager.yaml
```
