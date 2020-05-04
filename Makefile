OS_NAME := $(shell uname -s | tr A-Z a-z)
TAG := $(shell git rev-parse --short $(TRAVIS_COMMIT))
REPO := last9inc/k8stream

build_binary:
	env GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -o k8stream -installsuffix \
    		cgo github.com/last9/k8stream/

build: build_binary
	docker build -t $(REPO):latest .

test:
	go test -v ./...

docker_login:
	echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USERNAME)" --password-stdin

upload: docker_login
	docker tag $(REPO):latest $(REPO):$(TAG)
	docker push $(REPO):$(TAG)
	docker push $(REPO):latest
