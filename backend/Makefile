APP_NAME := exponia-api
BUILD_DIR := bin
MAIN_PACKAGE := ./cmd/api

.PHONY: build run clean

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PACKAGE)

run:
	go run $(MAIN_PACKAGE)

clean:
	rm -rf $(BUILD_DIR)