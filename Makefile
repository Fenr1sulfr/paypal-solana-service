# Makefile for a Go project

APP_NAME := RebaseDiplomaSystem
GO_FILES := $(shell find . -type f -name '*.go')
BUILD_DIR := build
BIN := $(BUILD_DIR)/$(APP_NAME)


# Путь к proto-файлам
PROTO_DIR=./proto
GEN_DIR=./internal/ports/grpc/pb
PROTO_FILE=$(PROTO_DIR)/payment.proto  # Full path for dependency
# Компилятор gRPC
PROTOC=protoc

# Генерация gRPC кода
proto: $(PROTO_FILE)
	$(PROTOC) -I $(PROTO_DIR) --go_out=$(GEN_DIR) --go_opt=paths=source_relative \
	--go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative $(PROTO_FILE)

proto2: 
	protoc --go_out=$(GEN_DIR) --go_opt=paths=source_relative  --go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative $(PROTO_FILE) 
	
build: $(GO_FILES)
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BIN) .

run: build
	@echo "Running $(APP_NAME)..."
	@$(BIN)

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "Cleaning up..."
	