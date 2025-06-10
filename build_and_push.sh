#!/bin/bash

# Check if version argument is passed
if [ $# -eq 0 ]; then
    echo "No version specified. Usage: $0 <version>"
    exit 1
fi

VERSION=$1

# Build the main Docker image
docker build -t quay.io/amdaecgt2/amd-t2:v$VERSION .
if [ $? -ne 0 ]; then
    echo "Failed to build the main Docker image."
    exit 1
fi

# Push the main Docker image
docker push quay.io/amdaecgt2/amd-t2:v$VERSION
if [ $? -ne 0 ]; then
    echo "Failed to push the main Docker image."
    exit 1
fi

# Generate kustomize manifests
operator-sdk generate kustomize manifests -q
if [ $? -ne 0 ]; then
    echo "Failed to generate kustomize manifests."
    exit 1
fi

# Build and generate the bundle
kustomize build config/manifests | operator-sdk generate bundle -q --overwrite --version $VERSION
if [ $? -ne 0 ]; then
    echo "Failed to generate the bundle."
    exit 1
fi

# Build the bundle Docker image
docker build -f bundle.Dockerfile -t quay.io/amdaecgt2/amd-t2-bundle:$VERSION .
if [ $? -ne 0 ]; then
    echo "Failed to build the bundle Docker image."
    exit 1
fi

# Push the bundle Docker image
docker push quay.io/amdaecgt2/amd-t2-bundle:$VERSION
if [ $? -ne 0 ]; then
    echo "Failed to push the bundle Docker image."
    exit 1
fi

echo "All commands executed successfully."

