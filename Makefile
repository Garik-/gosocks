include includes.mk

.PHONY: lint
lint: golangci-lint-install ## run golangci-linter
	@$(GOLANGCI_LINT_BIN) run ./...

.PHONY: run
run:
	@$(GO) run cmd/gosocks/main.go

.PHONY: client
client:
	curl -x socks5h://localhost:8080 https://google.com/