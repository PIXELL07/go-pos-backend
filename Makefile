.PHONY: run build test lint docker-up docker-down migrate

# Run dev server
run:
	go run ./cmd/server/main.go

# Build binary
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/pos-backend ./cmd/server

# Run tests
test:
	go test ./... -v -cover

# Lint
lint:
	golangci-lint run ./...

# Start infrastructure (postgres + redis)
infra-up:
	docker-compose up -d postgres redis

# Stop infrastructure
infra-down:
	docker-compose down

# Full docker stack
docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down -v

# Format code
fmt:
	gofmt -w .
	goimports -w .

# Tidy deps
tidy:
	go mod tidy

# Generate swagger docs (requires swag: go install github.com/swaggo/swag/cmd/swag@latest)
swagger:
	swag init -g cmd/server/main.go -o docs/swagger

# Watch mode (requires air: go install github.com/cosmtrek/air@latest)
dev:
	air -c .air.toml

help:
	@echo "Available commands:"
	@echo "  make run        - Run development server"
	@echo "  make build      - Build binary"
	@echo "  make infra-up   - Start Postgres + Redis"
	@echo "  make docker-up  - Full docker stack"
	@echo "  make test       - Run tests"
	@echo "  make dev        - Hot reload with Air"
