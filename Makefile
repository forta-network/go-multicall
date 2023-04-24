export GOBIN = $(shell pwd)/toolbin

LINT = $(GOBIN)/golangci-lint
FORMAT = $(GOBIN)/goimports

ABIGEN = $(GOBIN)/abigen

.PHONY: tools
tools:
	@echo 'Installing tools...'
	@rm -rf toolbin
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.2
	@go install golang.org/x/tools/cmd/goimports@v0.1.11

	@go install github.com/ethereum/go-ethereum/cmd/abigen@v1.11.5

.PHONY: require-tools
require-tools: tools
	@echo 'Checking installed tools...'
	@file $(LINT) > /dev/null
	@file $(FORMAT) > /dev/null

	@file $(ABIGEN) > /dev/null

	@echo "All tools found in $(GOBIN)!"

.PHONY: generate
generate: require-tools
	@$(ABIGEN) --out contracts/contract_multicall/multicall.go \
		--abi contracts/contract_multicall/abi.json --pkg contract_multicall \
		--type Multicall

.PHONY: test
test:
	go test -v -count=1 -covermode=count -coverprofile=coverage.out

.PHONY: cover
cover: test
	go tool cover -func=coverage.out -o=coverage.out

.PHONY: coverage
coverage: test
	go tool cover -func=coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}'
