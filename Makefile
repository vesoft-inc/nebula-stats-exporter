PROJECT="nebula-exporter"

GO ?= go
IMAGE_TAG ?= latest
DOCKER_REPO ?= reg.vesoft-inc.com/cloud-dev
IMAGE_REPOSITORY ?= $(DOCKER_REPO)/nebula-stats-exporter

export GO111MODULE := on
LDFLAGS = $(if $(DEBUGGER),,-s -w)
GOOS := $(if $(GOOS),$(GOOS),linux)
GOARCH := $(if $(GOARCH),$(GOARCH),amd64)
GOENV  := GO15VENDOREXPERIMENT="1" CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH)
GO     := $(GOENV) go
GO_BUILD := $(GO) build -trimpath
TARGETDIR := "$(GOOS)/$(GOARCH)"
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

PLATFORMS = arm64 amd64
BUILDX_PLATFORMS = linux/arm64,linux/amd64

ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: check build-exporter helm-chart image-multiarch clean

check: fmt vet lint

build-exporter:
	$(GO_BUILD) -ldflags '$(LDFLAGS)' -o bin/$(TARGETDIR)/nebula-stats-exporter main.go

helm-chart:
	helm repo index charts --url https://vesoft-inc.github.io/nebula-stats-exporter/charts
	helm package charts/nebula-exporter
	mv nebula-exporter-*.tgz charts/

image-multiarch:
	$(foreach PLATFORM,$(PLATFORMS), echo -n "$(PLATFORM)..."; GOARCH=$(PLATFORM) make build-exporter;)
	echo "Building and pushing nebula-exporter image... $(BUILDX_PLATFORMS)"
	docker buildx rm exporter || true
	docker buildx create --driver-opt network=host --use --name=exporter
	docker buildx build \
			--no-cache \
			--pull \
			--push \
			--progress plain \
			--platform $(BUILDX_PLATFORMS) \
			-t "${IMAGE_REPOSITORY}:${IMAGE_TAG}" .

clean:
	rm -rf nebula-stats-exporter
	docker rmi -f $(IMAGE_REPOSITORY):$(IMAGE_TAG)

fmt:
	go fmt .

vet:
	go vet ./...

lint:
	$(GOBIN)/golangci-lint run

.PHONY: all
