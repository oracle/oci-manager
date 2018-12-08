### Multi-Node Multi-Availaibility domain apache httpd chart
- This chart uses Custom Resources to model and add dependencies of OCI resources

requirements:
- oci-manager (set of controllers for oci resources)
url: https://github.com/oracle/oci-manager

usage:
- update values in values.yaml with change-me as default value
- customize other values such as availability_domains and instances_per_availability_domain

maintainers:
- Mike Schwankl  mike.schwankl@oracle.com
