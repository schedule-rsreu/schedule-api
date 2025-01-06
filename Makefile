.DEFAULT_GOAL = verify

BIN=bin

export GOBIN=$(CURDIR)/$(BIN)# for windows

#export GOBIN=$(PWD)/$(BIN) # for unix

$(BIN)/golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.4

$(BIN)/gotestsum:
	go install gotest.tools/gotestsum@v1.11.0

$(BIN)/goimports:
	go install golang.org/x/tools/cmd/goimports@latest

.PHONY: install
install: $(BIN)/golangci-lint  $(BIN)/goimports  $(BIN)/gotestsum  $(BIN)/tagalign

.PHONY: lint
lint:
	$(BIN)/golangci-lint run --config=.golangci.yml ./...

.PHONY: fix
fix:
	gofmt -s -w .

	$(BIN)/goimports -l -w .

	$(BIN)/golangci-lint run --config=.golangci.yml ./... --fix

.PHONY: test
test:
	$(BIN)/gotestsum ./... -race

.PHONY: build
build:
	go build -v -o main ./cmd/main.go

.PHONY: run
run:
	make swag
	make build
	.\main

.PHONY: swag
swag:

	swag init --parseDependency --parseInternal -g ./internal/http/handlers/router.go

.PHONY: d
d:
	docker compose up

.PHONY: prod
prod:
	docker compose -f docker-compose.yml -f docker-compose.traefik.yml up --build -d
