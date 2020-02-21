proto:
	protoc --go_out=plugins=grpc:. *.proto

build: proto
	go build -race .

test: proto
	go test -v ./...
