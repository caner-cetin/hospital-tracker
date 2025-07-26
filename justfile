# Hospital Tracker Justfile
# Run `just` to see available commands

default:
    @just --list
build:
    go build -o bin/hospital-tracker .
run:
    go run .
dev:
    air
install-deps:
    go install github.com/cosmtrek/air@latest
    go install github.com/swaggo/swag/cmd/swag@latest
    go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest
deps:
    go mod download
    go mod tidy
fmt:
    go fmt ./...

vet:
    go vet ./...

lint:
    golangci-lint run

check: fmt vet lint

test-unit:
    go test -v -race ./tests/unit/...

test-integration:
    go test -v -race ./tests/integration/...

test: test-unit test-integration

swagger:
    swag init

clean:
    rm -rf bin/
    go clean -testcache
setup: install-deps deps swagger
ci: deps check