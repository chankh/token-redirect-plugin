# Makefile for building the Go WASM application and Docker image

# Variables
GCP_PROJECT_NAME = <GCP PROJECT NAME>
GCP_REGION = us-central1
PLUGIN_NAME = token-redirect-plugin
ACTION_NAME = token-redirect-action

# Go build settings
GOOS := wasip1
GOARCH := wasm
GO_BUILD_MODE := c-shared
GO_OUTPUT := plugin.wasm
GO_SOURCE := main.go

# Docker image settings
DOCKER_REPO_NAME := $(GCP_REGION)-docker.pkg.dev/$(GCP_PROJECT_NAME)
DOCKER_IMAGE_NAME := $(DOCKER_REPO_NAME)/$(PLUGIN_NAME)/wasm-plugin
DOCKER_IMAGE_TAG := $(shell git rev-parse --short HEAD)

# Build the WASM plugin
build:
	@echo "Building WASM plugin..."
	@mkdir -p out
	@env GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=$(GO_BUILD_MODE) -o $(GO_OUTPUT) $(GO_SOURCE)
	@echo "WASM plugin built successfully: $(GO_OUTPUT)"

# Run unit tests
test:
	@echo "Running tests..."
	@go test
	@echo "All tests ran successfully."

# Build the Docker image
docker-build: build
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .
	@echo "Docker image built successfully: $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)"

# Push the Docker image
docker-push:
	@echo "Pushing Docker image..."
	@docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
	@echo "Docker image pushed successfully: $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)"

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf out
	@echo "Cleaned up successfully."

# Create Google Cloud Artifact Registry Repo
gar:
	@echo "Creating repo on Google Cloud Artifact Registry..."
	@gcloud artifacts repositories create $(PLUGIN_NAME) --project $(GCP_PROJECT_NAME) --location=$(GCP_REGION) --repository-format=docker --description "Docker repo for $(PLUGIN_NAME) wasm plugin"
	@echo "Docker repo created successfully: $(DOCKER_REPO_NAME)/$(PLUGIN_NAME)"

# Create Service Extension plugin
wasm-plugin:
	@echo "Creating Service Extension Plugin..."
	@gcloud beta service-extensions wasm-plugins create $(PLUGIN_NAME) --image=$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) --main-version=git-$(DOCKER_IMAGE_TAG) --project $(GCP_PROJECT_NAME)
	@echo "Service Extension Plugin created successfully."

# Create Wasm action for the plugin
wasm-action:
	@echo "Creating Wasm Action..."
	@gcloud alpha service-extensions wasm-actions create $(ACTION_NAME) --wasm-plugin=$(PLUGIN_NAME) --project=$(GCP_PROJECT_NAME)

# Default target
.PHONY: all
all: test build docker-build

wasm: wasm-plugin wasm-action

.PHONY: test build docker-build docker-push clean
