#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ${GOPATH}/src/k8s.io/code-generator)}

# oci
${SCRIPT_ROOT}/hack/generate-groups.sh deepcopy \
  github.com/oracle/oci-manager/pkg/client github.com/oracle/oci-manager/pkg/apis \
  "ocicommon.oracle.com:v1alpha1"\
  --go-header-file ${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt

${SCRIPT_ROOT}/hack/generate-groups.sh all \
  github.com/oracle/oci-manager/pkg/client github.com/oracle/oci-manager/pkg/apis \
  "ocidb.oracle.com:v1alpha1 ocicore.oracle.com:v1alpha1 ocilb.oracle.com:v1alpha1 ocice.oracle.com:v1alpha1 ociidentity.oracle.com:v1alpha1 cloud.k8s.io:v1alpha1"\
  --go-header-file ${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt

# cloud
# ${SCRIPT_ROOT}/hack/generate-groups.sh all \
#  github.com/oracle/oci-manager/pkg/client github.com/oracle/oci-manager/pkg/apis \
#  "cloud.k8s.io:v1alpha1"\
#  --go-header-file ${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt

echo "go fmt..."
go fmt github.com/oracle/oci-manager/pkg/client/... github.com/oracle/oci-manager/pkg/apis/...
