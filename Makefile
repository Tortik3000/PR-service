LOCAL_BIN := $(CURDIR)/bin
GOLANGCI_BIN := $(LOCAL_BIN)/golangci-lint
GOIMPORTS_BIN := $(LOCAL_BIN)/goimports
GO_TEST=$(LOCAL_BIN)/gotest
GO_TEST_ARGS=-race -v -tags=integration_test ./...

all: generate lint test build

.PHONY: lint
lint:
	echo 'Running linter on files...'
	$(GOLANGCI_BIN) run \
	--config=.golangci.yaml \
	--sort-results \
	--max-issues-per-linter=0 \
	--max-same-issues=0

generate: bin-deps .generate

.PHONY: test
test:
	echo 'Running tests...'
	${GO_TEST} ${GO_TEST_ARGS}

bin-deps: .bin-deps
.bin-deps: export GOBIN := $(LOCAL_BIN)
.bin-deps: .create-bin
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@1.64.8 && \
	GOBIN=$(LOCAL_BIN) go install github.com/rakyll/gotest@v0.0.6 && \
	GOBIN=$(LOCAL_BIN) go install go.uber.org/mock/mockgen@latest && \
	mv $(LOCAL_BIN)/mockgen $(LOCAL_BIN)/mockgen_uber


.create-bin:
	rm -rf ./bin
	mkdir -p ./bin

.generate:
	$(info Generating code...)


	(PATH="$(PATH):$(LOCAL_BIN)" && go generate ./...)
	go mod tidy

build:
	go mod tidy
	go build -o ./bin/pr-service ./cmd/pr-service/