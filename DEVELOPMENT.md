To test and build oci-manager run:
```bash
make
```

To run the oci-manager from your local development environment run:

```bash
export KUBECONFIG=~/.kube/config
export OCICONFIG=~/.oci/config
make run
```

To deploy and run oci-manager in your Kubernetes development cluster run:
```bash
export OCICONFIG=~/.oci/config
export DOCKER_REGISTRY=k8sfed/oci-manager
make deploy
```

To create a docker image run:
```bash
make image
```
