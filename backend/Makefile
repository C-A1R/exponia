APP_NAME := exponia-api
BUILD_DIR := bin
MAIN_PACKAGE := ./cmd/api
ENV_FILE := ../.env

.PHONY: build run start test tidy clean

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PACKAGE)

run:
	set -a; \
	. $(ENV_FILE); \
	set +a; \
	go run $(MAIN_PACKAGE)

start: build
	set -a; \
	. $(ENV_FILE); \
	set +a; \
	./$(BUILD_DIR)/$(APP_NAME)

test:
	go test ./...

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR)