###  OKE automation for network and policy prerequisites to use the OKE / oracle container engine
- This chart uses Custom Resources to model and add dependencies of OCI resources

requirements:
- oci-manager (set of controllers for oci resources)
url: https://github.com/oracle/oci-manager

usage:
- update values in values.yaml with change-me as default value
- customize other values such as availability_domains

maintainers:
- Mike Schwankl  mike.schwankl@oracle.com
