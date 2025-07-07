GOLANGCI_LINT=$(shell go env GOPATH)/bin/golangci-lint

lint:
	${GOLANGCI_LINT} run --fix --timeout 120s ./...

