
-include .env
export $(shell sed 's/=.*//' .env)

APP_NAME := booking-service
CMD_DIR := cmd/server
PROTO_DIR := proto
MIGRATIONS_DIR := migrations

DB_URL := postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}
LOCAL_BIN := $(shell go env GOPATH)/bin

# ──────────────────────────────────────────────
# Help
# ──────────────────────────────────────────────
.PHONY: help
help:
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "  Run & Build"
	@echo "    run              Run service locally"
	@echo "    build            Build Go binary to bin/"
	@echo ""
	@echo "  Test"
	@echo "    test             Run all tests"
	@echo "    test-v           Run all tests (verbose)"
	@echo "    test-cover       Run tests with coverage report"
	@echo "    test-usecase     Run only usecase tests"
	@echo "    test-handler     Run only handler/grpc tests"
	@echo ""
	@echo "  Code Quality"
	@echo "    vet              Run go vet"
	@echo "    lint             Run vet + build + test (full check)"
	@echo ""
	@echo "  Database"
	@echo "    migrate-up       Run migrations up"
	@echo "    migrate-create   Create new migration (name=xxx)"
	@echo ""
	@echo "  Proto"
	@echo "    proto            Generate gRPC & Gateway code"
	@echo "    install-deps     Install protoc plugins"
	@echo ""
	@echo "  Docker"
	@echo "    docker-build     Build Docker image"
	@echo "    up               Docker compose up"
	@echo "    down             Docker compose down"
	@echo ""

# ──────────────────────────────────────────────
# Run & Build
# ──────────────────────────────────────────────
.PHONY: run
run:
	go run ./$(CMD_DIR)

.PHONY: build
build:
	@mkdir -p bin
	go build -o bin/$(APP_NAME) ./$(CMD_DIR)

# ──────────────────────────────────────────────
# Test
# ──────────────────────────────────────────────
.PHONY: test
test:
	go test ./...

.PHONY: test-v
test-v:
	go test ./... -v

.PHONY: test-cover
test-cover:
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -func=coverage.out
	@rm -f coverage.out

.PHONY: test-usecase
test-usecase:
	go test ./internal/usecase/... -v

.PHONY: test-handler
test-handler:
	go test ./internal/handler/grpc/... -v

# ──────────────────────────────────────────────
# Code Quality
# ──────────────────────────────────────────────
.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	go vet ./...
	go build ./...
	go test ./...

# ──────────────────────────────────────────────
# Database
# ──────────────────────────────────────────────
.PHONY: migrate-up
migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

.PHONY: migrate-create
migrate-create:
	@mkdir -p $(MIGRATIONS_DIR)
	@goose -dir $(MIGRATIONS_DIR) create $(name) sql

# ──────────────────────────────────────────────
# Proto
# ──────────────────────────────────────────────
.PHONY: install-deps
install-deps:
	@mkdir -p $(LOCAL_BIN)
	@GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@GOBIN=$(LOCAL_BIN) go install -mod=mod google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

.PHONY: vendor-proto
vendor-proto:
	@if [ ! -d vendor.protogen/google/api ]; then \
		echo "Cloning googleapis..."; \
		git clone --depth=1 https://github.com/googleapis/googleapis.git vendor.protogen/googleapis; \
		mkdir -p vendor.protogen/google; \
		mv vendor.protogen/googleapis/google vendor.protogen/; \
		rm -rf vendor.protogen/googleapis; \
	fi

.PHONY: proto
proto:
	protoc -I proto -I vendor.protogen \
		--go_out=$(PROTO_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PROTO_DIR) --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=$(PROTO_DIR) \
		$(PROTO_DIR)/*.proto

# ──────────────────────────────────────────────
# Docker
# ──────────────────────────────────────────────
.PHONY: docker-build
docker-build:
	docker build -t $(APP_NAME) .

.PHONY: up
up:
	docker compose up --build -d

.PHONY: down
down:
	docker compose down
