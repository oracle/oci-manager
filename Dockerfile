# build stage
FROM golang:alpine AS build-env
RUN apk update && apk add alpine-sdk
ENV src /go/src/github.com/oracle/oci-manager
ADD . ${src}
ARG VERSION=dirty
RUN export VERSION=${VERSION} && cd ${src} && make build

# final stage
FROM alpine
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-env /go/src/github.com/oracle/oci-manager/LICENSE /LICENSE
COPY --from=build-env /go/src/github.com/oracle/oci-manager/LICENSES.deps /LICENSES.deps
COPY --from=build-env /go/src/github.com/oracle/oci-manager/bin/oci-manager /oci-manager
ENTRYPOINT ["/oci-manager"]
