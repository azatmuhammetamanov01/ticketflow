-include .env
export $(shell sed 's/=.*//' .env)

APP_NAME := event-service
CMD_DIR := ./cmd/server
PROTO_DIR := proto
GOOGLEAPIS_DIR := vendor.protogen/googleapis
MIGRATIONS_DIR := migrations
BIN_DIR := bin
DB_URL := postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}
LOCAL_BIN := $(shell go env GOPATH)/bin

install-deps:
	@mkdir -p $(LOCAL_BIN)
	@GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@GOBIN=$(LOCAL_BIN) go install -mod=mod google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

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
	make -C .. proto

.PHONY: test
test:
	go test ./...

build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/server/main.go


run:
	go run ./$(CMD_DIR)

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

docker-build:
	docker build -t $(APP_NAME) .

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f
