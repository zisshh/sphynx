#!/bin/bash

# Generate unique image name using timestamp and random string
UNIQUE_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')
IMAGE_NAME="loadbalancer-tests-s3-${UNIQUE_ID,,}"

# Clean up from any previous runs
echo "Cleaning up previous runs..."
rm -rf solution __pycache__ test_results.json

# Copy solution directory to tests directory
echo "Copying solution files..."
cp -r ../solution .

# Remove any existing Docker image with the same name
echo "Removing existing Docker image if present..."
docker rmi -f $IMAGE_NAME > /dev/null 2>&1 || true

# Build the Docker image with unique name
echo "Building Docker image..."
docker build --no-cache -t $IMAGE_NAME .

# Run the container using the unique image and capture output
echo "Running tests with image: $IMAGE_NAME..."
CONTAINER_ID=$(docker run -d $IMAGE_NAME)

# Stream logs while waiting for container to finish
docker logs -f $CONTAINER_ID &
LOGS_PID=$!

# Wait for container to finish
EXIT_CODE=$(docker wait $CONTAINER_ID)
kill $LOGS_PID 2>/dev/null

# Add a small delay before trying to copy
sleep 2

# Try to copy test results multiple times
MAX_ATTEMPTS=3
for i in $(seq 1 $MAX_ATTEMPTS); do
    echo "Attempt $i to copy test results..."
    if docker cp $CONTAINER_ID:/app/test_results.json .; then
        break
    fi
    sleep 2
done

# Cleanup: Remove the container and image
echo "Cleaning up Docker resources..."
docker rm $CONTAINER_ID
docker rmi $IMAGE_NAME

# Cleanup local files
echo "Cleaning up local files..."
rm -rf solution __pycache__

# Check if test_results.json exists locally
if [ -f "test_results.json" ]; then
    echo "Tests completed and results copied successfully"
    exit $EXIT_CODE
else
    echo "Failed to get test results"
    exit 1
fi
