# git-write-mcp — local developer workflow (CI runs equivalent checks in GitHub Actions).

.PHONY: build test vet fmt check-fmt lint ci

BIN := bin/git-write-mcp

build:
	mkdir -p bin
	go build -trimpath -o $(BIN) ./cmd/git-write-mcp

test:
	go test -count=1 ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

check-fmt:
	test -z "$$(gofmt -l .)"

lint:
	golangci-lint run ./...

ci: check-fmt vet test lint build
