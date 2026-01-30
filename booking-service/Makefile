
-include .env
export $(shell sed 's/=.*//' .env)

APP_NAME := booking-service
CMD_DIR := cmd/booking-service
PROTO_DIR := proto
GOOGLEAPIS_DIR := googleapis
MIGRATIONS_DIR := migrations

DB_URL := postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}
LOCAL_BIN := $(shell go env GOPATH)/bin
install-deps:
	@mkdir -p $(LOCAL_BIN)
	@GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@GOBIN=$(LOCAL_BIN) go install -mod=mod google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

vendor-proto:
	@if [ ! -d vendor.protogen/google/api ]; then \
		echo "Cloning googleapis..."; \
		git clone --depth=1 https://github.com/googleapis/googleapis.git vendor.protogen/googleapis; \
		mkdir -p vendor.protogen/google; \
		mv vendor.protogen/googleapis/google vendor.protogen/; \
		rm -rf vendor.protogen/googleapis; \
	fi

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make proto            Generate gRPC & Gateway code"
	@echo "  make build            Build Go binary"
	@echo "  make run              Run service locally"
	@echo "  make migrate-create   Create new migration (name=xxx)"
	@echo "  make docker-build     Build Docker image"
	@echo "  make up               Docker compose up"

.PHONY: proto
proto:
	protoc -I $`(PROTO_DIR) -I $(GOOGLEAPIS_DIR) \
		--go_out=$(PROTO_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PROTO_DIR) --grpc-gateway_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto

.PHONY: build
build:
	go build -o bin/$(APP_NAME) ./$(CMD_DIR)

.PHONY: run
run:
	go run ./$(CMD_DIR)

.PHONY: migrate-create
migrate-create:
	@mkdir -p $(MIGRATIONS_DIR)
	@goose -dir $(MIGRATIONS_DIR) create $(name) sql

.PHONY: migrate-up
migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

.PHONY: docker-build
docker-build:
	docker build -t $(APP_NAME) .

.PHONY: up
up:
	docker compose up --build -d

.PHONY: down
down:
	docker compose down
