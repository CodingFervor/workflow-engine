.PHONY: build run test clean docker-up docker-down docker-build docker-run deps lint fmt vet coverage

APP_NAME=workflow-engine
BUILD_DIR=./bin
CONFIG_PATH=./configs/config.yaml
DOCKER_IMAGE=workflow-engine
DOCKER_TAG=latest

build:
	go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/api

run:
	go run ./cmd/api

test:
	go test -v -race ./...

clean:
	rm -rf $(BUILD_DIR) coverage.out

deps:
	go mod tidy

fmt:
	gofmt -w .

vet:
	go vet ./...

lint: fmt vet

coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run:
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-push:
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest
