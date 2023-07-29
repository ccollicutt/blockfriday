# File: Makefile

GO := go
BINARY_NAME := admission-controller
CERTS_DIR := certs
BUILD_DIR := build
# Set the REGISTRY with the ?= operator, so that it can be overridden by an environment variable
REGISTRY ?= someregistry
IMAGE_NAME := blockfriday:latest

.PHONY: all build clean test

all: image push rollout

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) .

deps:
	go mod tidy

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)

image:
	@echo "Building docker image ${REGISTRY}/${IMAGE_NAME}..."
	@docker build -t ${REGISTRY}/$(IMAGE_NAME) .

push:
	@echo "Pushing docker image ${REGISTRY}/${IMAGE_NAME}..."
	@docker push ${REGISTRY}/$(IMAGE_NAME)

rollout:
	@echo "Rolling out deployment..."
	@kubectl rollout restart deployment/blockfriday

redo-test:
	@echo "Redoing test..."
	-@kubectl delete -f test/deployment.yaml
	@kubectl create -f test/deployment.yaml

test:
	go test
