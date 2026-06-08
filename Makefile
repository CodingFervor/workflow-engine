.PHONY: build run test clean docker-up docker-down deps lint

APP_NAME=workflow-engine
BUILD_DIR=./bin
CONFIG_PATH=./configs/config.yaml

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/api

run:
	go run ./cmd/api

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR)

deps:
	go mod tidy

lint:
	go vet ./...

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
