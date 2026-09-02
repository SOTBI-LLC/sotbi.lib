DIRECTORY_PROTOC_VERSION := 36.0
DIRECTORY_PROTOC_GEN_GO_VERSION := v1.36.6
DIRECTORY_PROTOC_GEN_GO_GRPC_VERSION := v1.5.1
DIRECTORY_PROTOC_GEN_VALIDATE_VERSION := v1.2.1

gen_proto:
	protoc --go_out=./pkg/ --go_opt=paths=source_relative ./api/notification/*.proto
	protoc --go_out=./pkg/ --go_opt=paths=source_relative --go-grpc_out=./pkg/ --go-grpc_opt=paths=source_relative api/excel_gen/*.proto
	protoc --go_out=./pkg/ --go_opt=paths=source_relative --go-grpc_out=./pkg/ --go-grpc_opt=paths=source_relative api/onec/*.proto
	protoc -I . -I ./api --go_out=./pkg --go_opt=paths=source_relative \
	--go-grpc_out=./pkg --go-grpc_opt=paths=source_relative \
	--validate_out="lang=go,paths=source_relative:./pkg" ./api/motivation/v1/service.proto
	$(MAKE) gen_directory_v1
	protoc --go_out=./pkg/api --go_opt=paths=source_relative --go-grpc_out=./pkg/api --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=./pkg/api --grpc-gateway_opt=paths=source_relative -I ./api \
	--openapiv2_out=./pkg/api ./api/counterparty/*.proto

.PHONY: check_directory_proto_toolchain
check_directory_proto_toolchain:
	@test "$$(protoc --version)" = "libprotoc $(DIRECTORY_PROTOC_VERSION)" || \
		(echo "Directory v1 requires libprotoc $(DIRECTORY_PROTOC_VERSION), got $$(protoc --version)"; exit 1)

.PHONY: gen_directory_v1
gen_directory_v1: check_directory_proto_toolchain
	protoc -I . -I ./api --go_out=./pkg --go_opt=paths=source_relative \
	--go-grpc_out=./pkg --go-grpc_opt=paths=source_relative \
	--validate_out="lang=go,paths=source_relative:./pkg" ./api/directory/v1/service.proto

.PHONY: check_directory_codegen
check_directory_codegen: gen_directory_v1
	@git diff --exit-code -- api/directory/v1 pkg/api/directory/v1
