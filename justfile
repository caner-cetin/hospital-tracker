# Hospital Tracker Justfile
# Run `just` to see available commands

default:
    @just --list
build:
    go build -o bin/hospital-tracker .
run:
    go run .
dev:
    go run .
install-deps:
    go install github.com/swaggo/swag/cmd/swag@latest
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
deps:
    go mod download
    go mod tidy
fmt:
    go fmt ./...

vet:
    go vet ./...

lint:
    golangci-lint run --timeout=5m

check: fmt vet

test-unit:
    go test -v -race ./tests/unit/...

test-integration:
    go test -v -race ./tests/integration/...

test: test-unit test-integration

swagger:
    swag init

docker-build:
    docker build -t hospital-tracker .

clean:
    rm -rf bin/
    go clean -testcache
setup: install-deps deps swagger
ci: deps check