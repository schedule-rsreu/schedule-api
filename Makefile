.DEFAULT_GOAL = verify

BIN=bin

export GOBIN=$(CURDIR)/$(BIN)# for windows

#export GOBIN=$(PWD)/$(BIN) # for unix

$(BIN)/golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.3

$(BIN)/gotestsum:
	go install gotest.tools/gotestsum@v1.11.0

$(BIN)/goimports:
	go install golang.org/x/tools/cmd/goimports@latest

$(BIN)/tagalign:
	go install github.com/4meepo/tagalign/cmd/tagalign@latest

.PHONY: install
install: $(BIN)/golangci-lint  $(BIN)/goimports  $(BIN)/gotestsum  $(BIN)/tagalign

.PHONY: lint
lint:
	$(BIN)/goimports -l .

	"$(BIN)/tagalign" -sort -order "json,xml" -strict ./...

	$(BIN)/golangci-lint run --config=.golangci.yml ./...

.PHONY: fix
fix:
	gofmt -s -w .

	$(BIN)/goimports -l -w .

	"$(BIN)/tagalign" -fix -sort -order "json,xml" -strict ./...

	$(BIN)/golangci-lint run --config=.golangci.yml ./... --fix

.PHONY: ve
ve:
	bin\tagalign.exe -fix -sort -order "json,xml" -strict ./...

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
