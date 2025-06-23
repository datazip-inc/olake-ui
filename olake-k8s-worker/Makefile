.PHONY: build push run clean test

# Variables
IMAGE_NAME = olakego/olake-k8s-worker
TAG = latest
DOCKER_REGISTRY = docker.io

# Build the Docker image
build:
	@echo "Building Docker image..."
	docker build -t $(IMAGE_NAME):$(TAG) .

# Push the image to registry
push: build
	@echo "Pushing Docker image to registry..."
	docker push $(IMAGE_NAME):$(TAG)

# Run the container locally for testing
run:
	@echo "Running container locally..."
	docker run --rm -p 8080:8080 \
		-e TEMPORAL_ADDRESS=localhost:7233 \
		-e LOG_LEVEL=debug \
		$(IMAGE_NAME):$(TAG)

# Clean up local images
clean:
	@echo "Cleaning up Docker images..."
	docker rmi $(IMAGE_NAME):$(TAG) || true

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Build for specific platform (useful for ARM Macs)
build-multiarch:
	@echo "Building multi-architecture image..."
	docker buildx build --platform linux/amd64,linux/arm64 \
		-t $(IMAGE_NAME):$(TAG) \
		--push .

# Build and push in one command
deploy: test build push
	@echo "Build and push completed!"

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build Docker image"
	@echo "  push          - Push image to registry"
	@echo "  run           - Run container locally"
	@echo "  clean         - Remove local images"
	@echo "  test          - Run tests"
	@echo "  build-multiarch - Build for multiple architectures"
	@echo "  deploy        - Test, build and push"
	@echo "  help          - Show this help"
