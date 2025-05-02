gen_proto:
	protoc --go_out=./pkg/ --go_opt=paths=source_relative ./api/notification/*.proto
	protoc --go_out=./pkg/ --go_opt=paths=source_relative --go-grpc_out=./pkg/ --go-grpc_opt=paths=source_relative api/excel_gen/*.proto
	protoc --go_out=./pkg/api --go_opt=paths=source_relative --go-grpc_out=./pkg/api --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=./pkg/api --grpc-gateway_opt=paths=source_relative -I ./api \
	--openapiv2_out=./pkg/api ./api/counterparty/*.proto
