PROJECT="nebula-exporter"

GO ?= go
VERSION ?= v0.0.2
DockerUser=vesoft

ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: check build push clean

check: fmt vet lint

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build nebula-stats-exporter.go
	docker build -t $(DockerUser)/nebula-stats-exporter:$(VERSION) .

push:
	docker push $(DockerUser)/nebula-stats-exporter:$(VERSION)

clean:
	rm -rf nebula-stats-exporter
	docker rmi -f $(DockerUser)/nebula-stats-exporter:$(VERSION)

fmt:
	go fmt .

vet:
	go vet ./...

lint:
	$(GOBIN)/golangci-lint run

.PHONY: all



