GIT_COMMIT?=$(shell git rev-parse --short HEAD)
GIT_VERSION?=$(shell git describe --tags 2>/dev/null || echo "v0.0.0-"$(GIT_COMMIT))

export ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
BUILD_DIR:=bin

LDFLAGS := -w -s
LDFLAGS += -X "github.com/fgiudici/ddflare/pkg/version.Version=${GIT_VERSION}"


.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/ddflare

.PHONY: clean
clean:
	@rm -rf $(BUILD_DIR)