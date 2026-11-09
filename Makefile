APP_NAME := docker-tui
BIN_DIR := bin
CMD_PATH := ./cmd/docker-tui

.PHONY: all build run run-tui run-web test clean tidy

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo "Build complete: $(BIN_DIR)/$(APP_NAME)"

run: build
	./$(BIN_DIR)/$(APP_NAME)

run-tui: build
	./$(BIN_DIR)/$(APP_NAME) tui

run-web: build
	./$(BIN_DIR)/$(APP_NAME) web

test:
	go test ./...

tidy:
	go mod tidy

clean:
	rm -rf $(BIN_DIR)
