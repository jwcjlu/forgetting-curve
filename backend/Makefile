.PHONY: init
init:
	go get -u github.com/go-kratos/kratos/cmd/kratos/v2
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/google/wire/cmd/wire@latest

.PHONY: api
api:
	cd api && find . -name "*.proto" | xargs -I {} protoc --proto_path=. \
		--proto_path=../third_party \
		--go_out=paths=source_relative:. \
		--go-http_out=paths=source_relative:. \
		--go-grpc_out=paths=source_relative:. \
		{}

.PHONY: wire
wire:
	cd cmd/server && wire

.PHONY: build
build:
	go build -ldflags "-X main.Version=$(VERSION)" -o bin/server ./cmd/server

.PHONY: run
run:
	go run ./cmd/server -conf ../../configs

.PHONY: test
test:
	go test -v ./... -cover



