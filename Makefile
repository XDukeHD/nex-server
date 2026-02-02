.PHONY: build run clean

BINARY_NAME=nex-server
BUILD_DIR=bin

build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/nex-server

run: build
	@echo "Running..."
	@sudo $(BUILD_DIR)/$(BINARY_NAME)

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
