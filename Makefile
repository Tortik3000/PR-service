LOCAL_BIN := $(CURDIR)/bin
GOLANGCI_BIN := $(LOCAL_BIN)/golangci-lint

.PHONY: lint
lint:
	echo 'Running linter on files...'
	$(GOLANGCI_BIN) run \
	--config=.golangci.yaml \
	--sort-results \
	--max-issues-per-linter=0 \
	--max-same-issues=0

build:
	go mod tidy
	go build -o ./bin/pr-service ./cmd/pr-service/