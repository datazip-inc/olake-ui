.PHONY: build-local build-minikube run-local deploy-minikube deploy-full-stack test test-coverage clean logs status delete delete-full-stack port-forward-health port-forward-temporal minikube-services open-ui open-temporal format lint install-deps help

# Variables
BINARY_NAME = olake-k8s-worker
IMAGE_NAME = olakego/olake-k8s-worker
TAG = local
NAMESPACE = olake

# Build Go binary locally
build-local:
	@echo "Building Go binary locally..."
	go build -o $(BINARY_NAME) ./main.go

# Build Docker image and load into minikube
build-minikube:
	@echo "Building Docker image for minikube..."
	docker build -t $(IMAGE_NAME):$(TAG) .
	@echo "Loading image into minikube..."
	minikube image load $(IMAGE_NAME):$(TAG)

# Run worker locally (requires minikube cluster access)
run-local: build-local
	@echo "Running worker locally..."
	@echo "Make sure minikube is running and kubectl is configured"
	./$(BINARY_NAME)

# Deploy worker to minikube cluster
deploy-minikube: build-minikube
	@echo "Deploying worker to minikube..."
	kubectl apply -f k8s/manifests/olake-worker/
	@echo "Waiting for deployment to be ready..."
	kubectl wait --for=condition=available --timeout=300s deployment/olake-k8s-worker -n $(NAMESPACE)

# Deploy full stack to minikube
deploy-full-stack:
	@echo "Deploying full OLake stack to minikube..."
	cd k8s/manifests/full-stack/scripts && ./deploy-all.sh

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean up local artifacts
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	docker rmi $(IMAGE_NAME):$(TAG) 2>/dev/null || true

# Show worker logs from minikube
logs:
	@echo "Showing worker logs..."
	kubectl logs -l app.kubernetes.io/name=olake-k8s-worker -n $(NAMESPACE) -f --tail=100

# Check deployment status in minikube
status:
	@echo "Checking deployment status..."
	@echo "\n=== Pods ==="
	kubectl get pods -n $(NAMESPACE) -l app.kubernetes.io/name=olake-k8s-worker
	@echo "\n=== Services ==="
	kubectl get svc -n $(NAMESPACE) -l app.kubernetes.io/name=olake-k8s-worker
	@echo "\n=== Deployments ==="
	kubectl get deployments -n $(NAMESPACE) -l app.kubernetes.io/name=olake-k8s-worker

# Check full stack status
status-full:
	@echo "Checking full stack status..."
	cd k8s/manifests/full-stack/scripts && ./status.sh

# Delete worker from minikube
delete:
	@echo "Deleting worker from minikube..."
	kubectl delete -f k8s/manifests/olake-worker/ --ignore-not-found=true

# Delete full stack from minikube
delete-full-stack:
	@echo "Deleting full stack from minikube..."
	cd k8s/manifests/full-stack/scripts && ./cleanup.sh

# Port forward to worker health endpoint
port-forward-health:
	@echo "Port forwarding to worker health endpoint..."
	kubectl port-forward -n $(NAMESPACE) service/olake-k8s-worker 8090:8090

# Port forward to Temporal UI
port-forward-temporal:
	@echo "Port forwarding to Temporal UI..."
	kubectl port-forward -n $(NAMESPACE) service/temporal-ui 8080:8080

# Show minikube services
minikube-services:
	@echo "Showing minikube services..."
	minikube service list -n $(NAMESPACE)

# Open OLake UI in browser
open-ui:
	@echo "Opening OLake UI..."
	minikube service olake -n $(NAMESPACE) --url | head -1 | xargs -I {} sh -c 'open "{}:30082" || xdg-open "{}:30082"'

# Open Temporal UI in browser
open-temporal:
	@echo "Opening Temporal UI..."
	minikube service temporal-ui -n $(NAMESPACE) --url | xargs -I {} sh -c 'open "{}" || xdg-open "{}"'

# Format Go code
format:
	@echo "Formatting Go code..."
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Install development dependencies
install-deps:
	@echo "Installing development dependencies..."
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Help
help:
	@echo "OLake K8s Worker - Local Development"
	@echo ""
	@echo "Available targets:"
	@echo "  build-local       - Build Go binary locally"
	@echo "  build-minikube    - Build Docker image and load into minikube"
	@echo "  run-local         - Run worker locally (requires minikube access)"
	@echo "  deploy-minikube   - Deploy worker to minikube cluster"
	@echo "  deploy-full-stack - Deploy complete OLake stack to minikube"
	@echo "  test              - Run unit tests"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo "  clean             - Clean up local artifacts"
	@echo "  logs              - Show worker logs from minikube"
	@echo "  status            - Check worker deployment status"
	@echo "  status-full       - Check full stack status"
	@echo "  delete            - Delete worker from minikube"
	@echo "  delete-full-stack - Delete full stack from minikube"
	@echo "  port-forward-health - Port forward to health endpoint"
	@echo "  port-forward-temporal - Port forward to Temporal UI"
	@echo "  minikube-services - Show all minikube services"
	@echo "  open-ui           - Open OLake UI in browser"
	@echo "  open-temporal     - Open Temporal UI in browser"
	@echo "  format            - Format Go code"
	@echo "  lint              - Run linter"
	@echo "  install-deps      - Install development dependencies"
	@echo "  help              - Show this help"
	@echo ""
	@echo "Quick start:"
	@echo "  make deploy-full-stack  # Deploy everything to minikube"
	@echo "  make open-ui            # Open OLake UI"
	@echo "  make logs               # Monitor worker logs"
