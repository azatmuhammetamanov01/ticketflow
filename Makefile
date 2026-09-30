.PHONY: proto test build vendor-proto

GEN_DIR := gen
API_PROTO_DIR := api/proto
VENDOR_DIR := vendor.protogen

vendor-proto:
	@if [ ! -d $(VENDOR_DIR)/google/api ]; then \
		echo "Cloning googleapis..."; \
		git clone --depth=1 https://github.com/googleapis/googleapis.git $(VENDOR_DIR)/googleapis; \
		mkdir -p $(VENDOR_DIR)/google; \
		mv $(VENDOR_DIR)/googleapis/google $(VENDOR_DIR)/; \
		rm -rf $(VENDOR_DIR)/googleapis; \
	fi

proto: vendor-proto
	@mkdir -p $(GEN_DIR)/go
	protoc -I $(API_PROTO_DIR) -I $(VENDOR_DIR) \
		--go_out=$(GEN_DIR) --go_opt=module=github.com/azatmuhammetamanov01/online-ticket-booking/gen \
		--go-grpc_out=$(GEN_DIR) --go-grpc_opt=module=github.com/azatmuhammetamanov01/online-ticket-booking/gen \
		--grpc-gateway_out=$(GEN_DIR) --grpc-gateway_opt=module=github.com/azatmuhammetamanov01/online-ticket-booking/gen \
		$(API_PROTO_DIR)/event/v1/*.proto $(API_PROTO_DIR)/booking/v1/*.proto
	@echo "Protobuf code successfully generated in $(GEN_DIR)/go"

test:
	cd event-service && go test ./...
	cd booking-service && go test ./...

build:
	cd event-service && go build ./cmd/server
	cd booking-service && go build ./cmd/server
