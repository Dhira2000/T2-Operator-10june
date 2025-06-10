#!/bin/bash

# Check if version argument is passed
if [ $# -eq 0 ]; then
    echo "No version specified. Usage: $0 <version>"
    exit 1
fi

VERSION=$1

# Set Docker client timeout
export DOCKER_CLIENT_TIMEOUT=300
export COMPOSE_HTTP_TIMEOUT=300

# Create and use a new builder instance
docker buildx create --name mybuilder --use
docker buildx inspect --bootstrap

# Build and push the main Docker image for multiple architectures
docker buildx build --platform linux/amd64,linux/arm64 -t quay.io/amdaecgt2/amd-t2:v$VERSION --push .
if [ $? -ne 0 ]; then
    echo "Failed to build and push the main Docker image."
    docker buildx rm mybuilder
    exit 1
fi

# Generate kustomize manifests
operator-sdk generate kustomize manifests -q
if [ $? -ne 0 ]; then
    echo "Failed to generate kustomize manifests."
    docker buildx rm mybuilder
    exit 1
fi

# Build and generate the bundle
kustomize build config/manifests | operator-sdk generate bundle -q --overwrite --version $VERSION
if [ $? -ne 0 ]; then
    echo "Failed to generate the bundle."
    docker buildx rm mybuilder
    exit 1
fi

# Build and push the bundle Docker image for multiple architectures
docker buildx build --platform linux/amd64,linux/arm64 -f bundle.Dockerfile -t quay.io/amdaecgt2/amd-t2-bundle:$VERSION --push .
if [ $? -ne 0 ]; then
    echo "Failed to build and push the bundle Docker image."
    docker buildx rm mybuilder
    exit 1
fi

echo "All commands executed successfully."

# Remove the builder instance
docker buildx rm mybuilder
