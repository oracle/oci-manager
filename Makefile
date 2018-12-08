# Copyright 2018 Oracle and/or its affiliates. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DOCKER_REGISTRY ?= phx.ocir.io/k8sfed
TARGET := oci-manager
IMAGE:=${DOCKER_REGISTRY}/${TARGET}

BUILD := $(shell git describe --always --dirty)
# Allow overriding for release versions
# Else just equal the build (git hash)
VERSION ?= ${BUILD}

# directories which hold app source (not vendored or generated)
SRC_PKGS := ./cmd/... ./pkg/apis/... ./pkg/controller/...

.PHONY: all
all: test build

.PHONY: fmt
fmt:
	go fmt ${SRC_PKGS}

.PHONY: vet
vet:
	go vet ${SRC_PKGS}

.PHONY: test
test: fmt
	go test ${SRC_PKGS} -args -v=1 -logtostderr

.PHONY: build
build: fmt
	go build \
		-ldflags "-X main.version=${VERSION} -X main.build=${BUILD}" \
		-o "bin/${TARGET}" \
		./cmd/${TARGET}.go

.PHONY: run
run:
	go run ./cmd/${TARGET}.go

.PHONY: clean
clean:
	rm -rf bin

# .PHONY: deploy
# deploy:
# 	kubectl -n kube-system set image ds/${TARGET} ${TARGET}=${IMAGE}:${VERSION}
#

.PHONY: image
image:
	docker build --build-arg VERSION=${VERSION} -t ${IMAGE}:${VERSION} -f Dockerfile .

.PHONY: publish
publish:
	docker push ${IMAGE}:${VERSION}

.PHONY: vendor
vendor:
	glide install -v

.PHONY: license
license:
	hack/update-licenses-fileheaders.sh
	hack/update-licenses-deps.sh
