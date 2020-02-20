build:
	protoc --go_out=plugins=grpc:. *.proto
	go build -race .
