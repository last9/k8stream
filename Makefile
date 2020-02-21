proto:
	which protoc || ((which unzip || apt install unzip) && curl -sL https://github.com/protocolbuffers/protobuf/releases/download/v3.9.2/protoc-3.9.2-linux-x86_64.zip -o /tmp/protoc.zip && mkdir -p /tmp/protoc && unzip -o /tmp/protoc.zip -d /tmp/protoc && sudo cp /tmp/protoc/bin/protoc /usr/local/bin/protoc && rm -rf /tmp/protoc.zip)
	go get github.com/golang/protobuf/proto && go get github.com/golang/protobuf/protoc-gen-go
	protoc --go_out=plugins=grpc:. *.proto

build: proto
	go build -race .

test: proto
	go test -v ./...
