.PHONY: up down build test test-race lint fmt tidy

up:
	docker compose -f infra/docker-compose.yml up -d

down:
	docker compose -f infra/docker-compose.yml down

build:
	@mkdir -p bin
	@for svc in gateway; do go build -trimpath -o bin/$$svc ./apps/$$svc/cmd; done

test:
	go test ./...

test-race:
	CGO_ENABLED=1 go test -race ./...

lint:
	go vet ./...
	@test -z "$$(gofmt -l apps packages)" || { echo "gofmt:"; gofmt -l apps packages; exit 1; }

fmt:
	gofmt -w apps packages

tidy:
	go mod tidy
