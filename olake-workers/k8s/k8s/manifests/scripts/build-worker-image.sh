#!/bin/bash

# Build script for olake-worker-k8s and olake-ui Docker images
# Usage: ./build-worker-image.sh

set -e  # Exit on any error

# Configuration for Worker
WORKER_IMAGE_NAME="olakego/olake-worker-k8s"
WORKER_IMAGE_TAG="local"
WORKER_FULL_IMAGE_NAME="${WORKER_IMAGE_NAME}:${WORKER_IMAGE_TAG}"
WORKER_DOCKERFILE_PATH="Dockerfile"
WORKER_BUILD_CONTEXT="."

# Configuration for UI
UI_IMAGE_NAME="olake-ui"
UI_IMAGE_TAG="local"
UI_FULL_IMAGE_NAME="${UI_IMAGE_NAME}:${UI_IMAGE_TAG}"
UI_DOCKERFILE_PATH="Dockerfile"
UI_BUILD_CONTEXT="."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Get directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKER_DIR="$(cd "$SCRIPT_DIR/../../../" && pwd)"
UI_DIR="$(cd "$SCRIPT_DIR/../../../../" && pwd)"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running or not accessible"
    exit 1
fi

# Function to build an image
build_image() {
    local image_name=$1
    local full_image_name=$2
    local dockerfile_path=$3
    local build_context=$4
    local build_dir=$5
    
    print_status "=================================================="
    print_status "Building $full_image_name"
    print_status "=================================================="
    
    print_status "Navigating to directory: $build_dir"
    cd "$build_dir"
    
    # Check if we're in the correct directory
    if [ ! -f "$dockerfile_path" ]; then
        print_error "Dockerfile not found at $dockerfile_path"
        print_error "Expected to be in directory: $build_dir"
        return 1
    fi
    
    print_status "Build context: $build_context"
    print_status "Dockerfile: $dockerfile_path"
    
    # Build the Docker image
    print_status "Building Docker image..."
    if docker build -t "$full_image_name" -f "$dockerfile_path" "$build_context"; then
        print_success "Docker image built successfully!"
        print_success "Image: $full_image_name"
        
        # Show image details
        print_status "Image details:"
        docker images "$image_name" --format "table {{.Repository}}\t{{.Tag}}\t{{.ID}}\t{{.Size}}\t{{.CreatedAt}}" | head -2
        
        print_success "Build completed successfully!"
        return 0
    else
        print_error "Docker build failed!"
        return 1
    fi
}

# Build Worker Image
print_status "Starting Docker builds..."
if ! build_image "$WORKER_IMAGE_NAME" "$WORKER_FULL_IMAGE_NAME" "$WORKER_DOCKERFILE_PATH" "$WORKER_BUILD_CONTEXT" "$WORKER_DIR"; then
    print_error "Worker image build failed!"
    exit 1
fi

# Build UI Image  
if ! build_image "$UI_IMAGE_NAME" "$UI_FULL_IMAGE_NAME" "$UI_DOCKERFILE_PATH" "$UI_BUILD_CONTEXT" "$UI_DIR"; then
    print_error "UI image build failed!"
    exit 1
fi

print_success "All Docker images built successfully!"