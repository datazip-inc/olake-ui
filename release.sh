#!/usr/bin/env bash

# Function for colored output
function chalk() {
    local color=$1
    local text=$2
    local color_code=0
    if [[ $color == "red" ]]; then
        color_code=1
    elif [[ $color == "green" ]]; then
        color_code=2
    fi
    if [[ -n "$TERM" ]]; then
        echo -e "$(tput setaf $color_code)${text}$(tput sgr0)"
    else
        if [[ $color == "red" ]]; then
            echo -e "\033[31m${text}\033[0m"
        elif [[ $color == "green" ]]; then
            echo -e "\033[32m${text}\033[0m"
        else
            echo -e "${text}"
        fi
    fi
}

function fail() {
    local error="${1:-Unknown error}"
    echo "$(chalk red "${error}")"
    exit 1
}

function build_frontend() {
    echo "Building frontend application..."
    cd ui || fail "Failed to change to UI directory"
    echo "Installing dependencies in $(pwd)"
    pnpm install || fail "pnpm install failed"
    echo "Building application"
    pnpm run build || fail "pnpm build failed"
    cd - || fail "Failed to return to original directory"
    echo "$(chalk green "✅ Frontend successfully built")"
}

function setup_buildx() {
    echo "Setting up Docker buildx and QEMU..."
    docker buildx version >/dev/null 2>&1 || fail "Docker buildx is not installed. Please install it."
    docker run --rm --privileged multiarch/qemu-user-static --reset -p yes || fail "Failed to set up QEMU"
    docker buildx create --use --name multiarch-builder || echo "Buildx builder already exists, using it."
    docker buildx inspect --bootstrap || fail "Failed to bootstrap buildx builder"
    echo "✅ Buildx and QEMU setup complete"
}

# Function to release just the frontend
function release_frontend() {
    local version=$1
    local platform=$2
    local environment=$3
    local image_name="$DOCKER_REPO"
    
    # Set tag based on environment
    local tag_version=""
    local latest_tag=""
    
    case "$environment" in
        "master")
            tag_version="${version}"
            latest_tag="latest"
            ;;
        "staging")
            tag_version="stag-${version}"
            latest_tag="stag-latest"
            ;;
        *)
            tag_version="dev-${version}"
            latest_tag="dev-latest"
            ;;
    esac

    echo "Logging into Docker..."
    docker login -u="$DOCKER_LOGIN" -p="$DOCKER_PASSWORD" || fail "Docker login failed for $DOCKER_LOGIN"
    
    echo "**** Releasing frontend image $image_name for platforms [$platform] with version [$tag_version] ****"

    echo "Building and pushing frontend Docker image..."
    
    docker buildx build --platform "$platform" --push \
        -t "${image_name}:${tag_version}" \
        -t "${image_name}:${latest_tag}" \
        --build-arg ENVIRONMENT="$environment" \
        --build-arg APP_VERSION="$version" \
        -f Dockerfile . || fail "Frontend build failed. Exiting..."
    
    echo "$(chalk green "Frontend release successful for $image_name version $tag_version")"
}

# Function to release the temporal worker
function release_worker() {
    local version=$1
    local platform=$2
    local environment=$3 # Could be 'dev', 'staging', 'master', etc.
    local image_name="$DOCKER_REPO_WORKER" # Use a specific repo name for the worker, e.g., yourdockerhubuser/olake-worker
    
    # Set tag based on environment
    local tag_version=""
    local latest_tag=""
    
    case "$environment" in
        "master")
            tag_version="${version}"
            latest_tag="latest"
            ;;
        "staging")
            tag_version="stag-${version}"
            latest_tag="stag-latest"
            ;;
        "dev"|*) # Default to dev prefix if not master or staging
            tag_version="dev-${version}"
            latest_tag="dev-latest"
            ;;
    esac

    # # It's good practice to ensure DOCKER_REPO_WORKER is set
    # if [ -z "$DOCKER_REPO_WORKER" ]; then
    #     echo "$(chalk red "Error: DOCKER_REPO_WORKER environment variable is not set.")"
    #     return 1 # Or use fail "DOCKER_REPO_WORKER not set" if 'fail' is a global helper
    # fi

    echo "Logging into Docker (if not already logged in by a previous function call)..."
    # Assuming DOCKER_LOGIN and DOCKER_PASSWORD are set globally or passed
    # If login is handled globally at the start of the script, this might be redundant
    # but doesn't hurt to ensure.
    docker login -u="$DOCKER_LOGIN" -p="$DOCKER_PASSWORD" || fail "Docker login failed for $DOCKER_LOGIN"
    
    echo "**** Releasing worker image $image_name for platforms [$platform] with version [$tag_version] ****"

    echo "Building and pushing worker Docker image..."
    
    # Assuming worker.Dockerfile is in the project root (context '.')
    # If worker.Dockerfile or its context (e.g., server files) are elsewhere, adjust paths.
    docker buildx build --platform "$platform" --push \
        -t "${image_name}:${tag_version}" \
        -t "${image_name}:${latest_tag}" \
        --build-arg ENVIRONMENT="$environment" \
        --build-arg APP_VERSION="$version" \
        -f worker.Dockerfile . || fail "Worker build failed. Exiting..."
    
    echo "$(chalk green "Worker release successful for $image_name version $tag_version")"
}

SEMVER_EXPRESSION='v([0-9]+\.[0-9]+\.[0-9]+)$'
STAGING_VERSION_EXPRESSION='v([0-9]+\.[0-9]+\.[0-9]+)-[a-zA-Z0-9_.-]+'

echo "Release tool running..."
CURRENT_BRANCH=$(git branch --show-current)
echo "Building on branch: $CURRENT_BRANCH"
echo "Environment: $ENVIRONMENT"
echo "Fetching remote changes from git with git fetch"
git fetch origin "$CURRENT_BRANCH" >/dev/null 2>&1
GIT_COMMITSHA=$(git rev-parse HEAD | cut -c 1-8)
echo "Latest commit SHA: $GIT_COMMITSHA"

echo "Running checks..."

# Verify Docker login
docker login -u="$DOCKER_LOGIN" -p="$DOCKER_PASSWORD" >/dev/null 2>&1 || fail "❌ Docker login failed. Ensure DOCKER_LOGIN and DOCKER_PASSWORD are set."
echo "✅ Docker login successful"

# Version validation based on environment (default is dev with no restrictions)
if [[ -z "$VERSION" ]]; then
    fail "❌ Version not set. Empty version passed."
fi

# Validate version format based on environment
if [[ "$ENVIRONMENT" == "master" ]]; then
    [[ $VERSION =~ $SEMVER_EXPRESSION ]] || fail "❌ Version $VERSION does not match semantic versioning required for master (e.g., v1.0.0)"
    echo "✅ Version $VERSION matches semantic versioning for master"
elif [[ "$ENVIRONMENT" == "staging" ]]; then
    [[ $VERSION =~ $STAGING_VERSION_EXPRESSION ]] || fail "❌ Version $VERSION does not match staging version format (e.g., v1.0.0-rc1)"
    echo "✅ Version $VERSION matches format for staging"
else
    echo "✅ Flexible versioning allowed for development: $VERSION"
fi

# Setup buildx and QEMU
setup_buildx

build_frontend

platform="linux/amd64,linux/arm64"
echo "✅ Releasing frontend application for environment $ENVIRONMENT with version $VERSION on platforms: $platform"

chalk green "=== Releasing Olake Frontend application ==="
chalk green "=== Environment: $ENVIRONMENT ==="
chalk green "=== Release version: $VERSION ==="

# Call the frontend-only release function
release_frontend "$VERSION" "$platform" "$ENVIRONMENT"
release_worker "$VERSION" "$platform" "$ENVIRONMENT"

echo "$(chalk green "✅ Frontend release process completed successfully")" 