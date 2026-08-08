APP_NAME := exponia-api
BUILD_DIR := bin
MAIN_PACKAGE := ./cmd/api
ENV_FILE := ../.env
MIGRATIONS_DIR := $(CURDIR)/migrations
MIGRATE_IMAGE := migrate/migrate

.PHONY: build run start test tidy clean migrate-up migrate-down migrate-version migrate-force sqlc

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
	gotestsum --format testname -- ./... -count=1

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR)

migrate-up:
	set -a; \
	. $(ENV_FILE); \
	set +a; \
	docker run --rm \
		--network host \
		-v $(MIGRATIONS_DIR):/migrations \
		$(MIGRATE_IMAGE) \
		-path=/migrations \
		-database "$$DATABASE_URL" \
		up

migrate-down:
	set -a; \
	. $(ENV_FILE); \
	set +a; \
	docker run --rm \
		--network host \
		-v $(MIGRATIONS_DIR):/migrations \
		$(MIGRATE_IMAGE) \
		-path=/migrations \
		-database "$$DATABASE_URL" \
		down 1

migrate-version:
	set -a; \
	. $(ENV_FILE); \
	set +a; \
	docker run --rm \
		--network host \
		-v $(MIGRATIONS_DIR):/migrations \
		$(MIGRATE_IMAGE) \
		-path=/migrations \
		-database "$$DATABASE_URL" \
		version

migrate-force:
	set -a; \
	. $(ENV_FILE); \
	set +a; \
	docker run --rm \
		--network host \
		-v $(MIGRATIONS_DIR):/migrations \
		$(MIGRATE_IMAGE) \
		-path=/migrations \
		-database "$$DATABASE_URL" \
		force $(VER)

sqlc:
	sqlc generate