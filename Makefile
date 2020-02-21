OS_NAME := $(shell uname -s | tr A-Z a-z)
TAG := $(shell git rev-parse --short $(TRAVIS_COMMIT))
REPO := docker.pkg.github.com/last9/last9/last9

proto-linux:
	which protoc || ((which unzip || apt install unzip) && curl -sL https://github.com/protocolbuffers/protobuf/releases/download/v3.9.2/protoc-3.9.2-linux-x86_64.zip -o /tmp/protoc.zip && mkdir -p /tmp/protoc && unzip -o /tmp/protoc.zip -d /tmp/protoc && cp /tmp/protoc/bin/protoc $(GOPATH)/bin/protoc && rm -rf /tmp/protoc.zip)

proto: proto-$(OS_NAME)
	go get github.com/golang/protobuf/proto && go get github.com/golang/protobuf/protoc-gen-go
	protoc --go_out=plugins=grpc:. *.proto

build_binary: proto
	env GOOS=linux GARCH=amd64 CGO_ENABLED=0 GOCACHE=/tmp/gocache go build -o k8stream -a -installsuffix \
    		cgo github.com/last9/k8stream/

build: #build_binary
	docker build -t last9:latest .

test: proto
	go test -v ./...

docker_login:
	echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USERNAME)" --password-stdin docker.pkg.github.com

upload: docker_login
	docker tag last9:latest $(REPO):$(TAG)
	docker tag last9:latest $(REPO):latest
	docker tag last9:latest $(REPO):$(TAG)
	docker push $(REPO):latest
