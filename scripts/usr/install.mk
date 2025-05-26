add_proto:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest

install_gotools:
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2
	go install github.com/segmentio/golines@latest
	go install github.com/bombsimon/wsl/v4/cmd/wsl@latest
	go install github.com/daixiang0/gci@latest
	go install github.com/vektra/mockery/v2@v2.52.4

install: add_proto install_gotools
