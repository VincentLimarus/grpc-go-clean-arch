MODULE      := VincentLimarus/grpc-golang
GOBIN       := $(shell go env GOPATH)/bin
PROTOC      := "$(CURDIR)/tools/bin/protoc"

.PHONY: proto build run test cover vet lint fmt clean

proto:
	PATH="$(GOBIN):$(CURDIR)/tools/bin:$$PATH" $(PROTOC) -I proto \
		--go_out=. --go_opt=module=$(MODULE) \
		--go-grpc_out=. --go-grpc_opt=module=$(MODULE) \
		proto/product.proto

build:
	go build ./...

run:
	go run ./cmd/server

test:
	go test ./... -count=1

cover:
	go test ./... -cover

vet:
	go vet ./...

fmt:
	gofmt -w cmd internal

lint:
	golangci-lint run ./...

clean:
	rm -f server cmd/server/server