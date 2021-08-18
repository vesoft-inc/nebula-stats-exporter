PROJECT="nebula-exporter"

GO ?= go
VERSION ?= v0.0.5
DockerUser=vesoft

ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: check build docker-build docker-push clean

check: fmt vet lint

build:
	go build -o nebula-stats-exporter main.go

build-helm:
	helm repo index charts --url https://vesoft-inc.github.io/nebula-stats-exporter/charts
	helm package charts/nebula-exporter
	mv nebula-exporter-*.tgz charts/

docker-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o nebula-stats-exporter main.go
	docker build -t $(DockerUser)/nebula-stats-exporter:$(VERSION) .

docker-push:
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
