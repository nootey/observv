# sqlc is not installed globally, so it runs pinned through go run
SQLC := go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1

# Application (run `make db` first)
run:
	go run ./cmd/api

simulate:
	go run ./cmd/simulator

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
COMPOSE := -f ./docker-compose.yaml

up:
	docker compose $(COMPOSE) up -d --build

down:
	docker compose $(COMPOSE) down

db:
	docker compose $(COMPOSE) up -d db

logs:
	docker compose $(COMPOSE) logs -f api simulator

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
