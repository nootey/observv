# sqlc is not installed globally, so it runs pinned through go run
SQLC := go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1

# Exported, so docker compose gets it as a build arg
export VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X github.com/nootey/observv/pkg/version.Version=$(VERSION)

# Application (run `make db` and `make stack-up` first)
run:
	go run -ldflags "$(LDFLAGS)" ./cmd/api

simulate:
	go run -ldflags "$(LDFLAGS)" ./cmd/simulator

mock:
	mockery --config=.mockery.yaml

# Migrations (using Go wrapper)
migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-status:
	go run ./cmd/migrate status

migrate-create:
	go run ./cmd/migrate create $(NAME)

migrate-reset:
	go run ./cmd/migrate reset

migrate-fresh:
	go run ./cmd/migrate reset
	go run ./cmd/migrate up
	go run ./cmd/seed

# Seeding
seed:
	go run ./cmd/seed

# Docker
COMPOSE             := -f ./docker-compose.yaml
COMPOSE_LOCAL       := -f ./docker-compose.local.yaml
COMPOSE_STACK       := -f ./docker-compose.stack.yaml
COMPOSE_STACK_LOCAL := -f ./docker-compose.stack.local.yaml

# App and stack. Grafana runs on http://localhost:3000
up:
	docker compose $(COMPOSE) $(COMPOSE_LOCAL) $(COMPOSE_STACK) $(COMPOSE_STACK_LOCAL) up -d --build

# Stack only, for the api on the host
stack-up:
	docker compose $(COMPOSE_STACK) $(COMPOSE_STACK_LOCAL) up -d

# App and stack behind Traefik. Grafana runs on https://grafana.$DOMAIN
prod-up:
	docker compose $(COMPOSE) $(COMPOSE_STACK) --profile proxy up -d --build

down:
	docker compose $(COMPOSE) $(COMPOSE_STACK) --profile proxy down

db:
	docker compose $(COMPOSE) $(COMPOSE_LOCAL) up -d db

logs:
	docker compose $(COMPOSE) logs -f api simulator

# Chaos (needs the api on the host port, so not with prod-up)
CHAOS_URL := http://127.0.0.1:8080/api/chaos

chaos-on:
	curl -sS -X POST $(CHAOS_URL) -H 'Content-Type: application/json' -d '{"latency_ms":300,"error_rate":0.2}'

chaos-off:
	curl -sS -X POST $(CHAOS_URL) -H 'Content-Type: application/json' -d '{"latency_ms":0,"error_rate":0}'

# SQLC
sqlc:
	$(SQLC) generate

# Tests
test:
	go test -race -count=1 ./...

test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Lint
lint:
	golangci-lint run

lint-fix:
	gofmt -w .
	golangci-lint run --fix

tidy:
	go mod tidy
	go mod verify

# Pre push checklist
pre-push:
	mockery --config=.mockery.yaml
	@go build ./... && echo "build successful" || (echo "build failed" && exit 1)
	@golangci-lint run && echo "lint successful" || (echo "lint failed" && exit 1)
	@go test -race -count=1 ./... && echo "tests successful" || (echo "tests failed" && exit 1)
