PROJECT="nebula-exporter"

GO           ?= go
VERSION ?= v0.0.1
DockerUser=vesoft

all: build push clean

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build nebula-exporter.go
	docker build -t $(DockerUser)/nebula-stats-exporter:$(VERSION) .

push:
	docker push $(DockerUser)/nebula-stats-exporter:$(VERSION)

clean:
	rm -rf nebula-exporter
	docker rmi -f $(DockerUser)/nebula-stats-exporter:$(VERSION)

fmt:
	go fmt .
.PHONY: all



