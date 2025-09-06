.DEFAULT_GOAL = verify

BIN=bin

export GOBIN=$(CURDIR)/$(BIN)# for windows


$(BIN)/golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

$(BIN)/gotestsum:
	go install gotest.tools/gotestsum@latest

$(BIN)/goimports:
	go install golang.org/x/tools/cmd/goimports@latest

.PHONY: install
install: $(BIN)/golangci-lint  $(BIN)/goimports  $(BIN)/gotestsum

.PHONY: lint
lint:
	$(BIN)/golangci-lint run --config=.golangci.yml ./...

.PHONY: format
format:
	gofmt -s -w .

	$(BIN)/goimports -l -w .

	$(BIN)/golangci-lint run --config=.golangci.yml ./... --fix

.PHONY: test
test:
	$(BIN)/gotestsum ./... -race -v -coverprofile=cover.out -covermode=atomic

.PHONY: build
build:
	go build -v -o main ./cmd/main.go

.PHONY: run
run:
	make swag
	make build
	.\main

.PHONY: brun
brun:
	go run -v ./cmd/main.go

.PHONY: run-main-file
run-main-file:
	.\main

.PHONY: swag
swag:

	swag init --parseDependency --parseInternal -g ./internal/http/handlers/router.go

.PHONY: d
d:
	docker compose up

.PHONY: testcover
testcover:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: prod
prod:
	docker compose -f docker-compose.yml -f docker-compose.traefik.yml up --build -d
