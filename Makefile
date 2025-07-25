GIT_COMMIT?=$(shell git rev-parse --short HEAD)
GIT_VERSION?=$(shell git describe --tags 2>/dev/null || echo "v0.0.0-"$(GIT_COMMIT))
IMG_REG?=ghcr.io
IMG_REPO?=ddflare/ddflare
IMG_TAG?=$(GIT_VERSION:v%=%)

export ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
BUILD_DIR   := bin
RELEASE_DIR := release

LDFLAGS := -w -s
LDFLAGS += -X "github.com/ddflare/ddflare/pkg/version.Version=${GIT_VERSION}"

COVERFILE?=coverage.out

GOOS   := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

.PHONY: build
build: build-cli

.PHONY: build-cli
build-cli:
	CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/ddflare  $(ROOT_DIR)/cli

.PHONY: clean
clean:
	@rm -vrf $(BUILD_DIR) $(COVERFILE) ${RELEASE_DIR}

.PHONY: release
release:
	$(call cross_compile,${GOOS},${GOARCH})

.PHONY: release-all
release-all: release-linux-amd64 release-linux-arm64 \
	release-linux-armv6 release-linux-armv7 \
	release-darwin-amd64 release-darwin-arm64 \
	release-windows-amd64 release-windows-arm64

.PHONY: release-linux-amd64
release-linux-amd64:
	$(call cross_compile,"linux","amd64","","")

.PHONY: release-linux-arm64
release-linux-arm64:
	$(call cross_compile,"linux","arm64","","")

.PHONY: release-linux-armv6
release-linux-armv6:
	$(call cross_compile,"linux","arm","6","-v6")

.PHONY: release-linux-armv7
release-linux-armv7:
	$(call cross_compile,"linux","arm","7","-v7")

.PHONY: release-darwin-amd64
release-darwin-amd64:
	$(call cross_compile,"darwin","amd64","","")

.PHONY: release-darwin-arm64
release-darwin-arm64:
	$(call cross_compile,"darwin","arm64","","")

.PHONY: release-windows-amd64
release-windows-amd64:
	$(call cross_compile,"windows","amd64","",".exe")

.PHONY: release-windows-arm64
release-windows-arm64:
	$(call cross_compile,"windows","arm64","",".exe")


.PHONY: docker
docker:
	DOCKER_BUILDKIT=1 docker build \
		-f Dockerfile \
		--build-arg "VERSION=${GIT_VERSION}" \
		-t ${IMG_REG}/${IMG_REPO}:${IMG_TAG}

.PHONY: unit-tests
unit-tests:
	@go test -coverprofile $(COVERFILE) ./pkg/...


define cross_compile
	mkdir -p ${RELEASE_DIR}
	$(eval $@_GOOS = $(1))
	$(eval $@_GOARCH = $(2))
	$(eval $@_GOARM = $(3))
	$(eval $@_SUFFIX = $(4))

	env GOOS=${$@_GOOS} GOARCH=${$@_GOARCH} GOARM=${$@_GOARM} CGO_ENABLED=0 go build \
		-ldflags '$(LDFLAGS)' -o ${RELEASE_DIR}/ddflare-${$@_GOOS}-${$@_GOARCH}${$@_SUFFIX} ${ROOT_DIR}/cli
	shasum -a 256 ${RELEASE_DIR}/ddflare-${$@_GOOS}-${$@_GOARCH}${$@_SUFFIX} \
		> ${RELEASE_DIR}/ddflare-${$@_GOOS}-${$@_GOARCH}${$@_SUFFIX}.sha256
endef
