#!/bin/bash

# Function to check if Go server is ready
check_go_server() {
    curl -s http://localhost:8080/health >/dev/null
    return $?
}

# Function to cleanup processes
cleanup() {
    # Only kill processes if they exist and are running
    if ps -p $GO_PID > /dev/null 2>&1; then
        kill $GO_PID 2>/dev/null
    fi
    if ps -p $BACKEND_PID > /dev/null 2>&1; then
        kill $BACKEND_PID 2>/dev/null
    fi
}

# Set trap for cleanup
trap cleanup EXIT

# Start the Go application
echo "Starting Go server..." >&2
cd /app/solution
go run main.go >&2 &
GO_PID=$!

# Wait for Go server to start with timeout
timeout=60
counter=0
echo "Waiting for Go server to start..." >&2
while ! check_go_server; do
    if [ $counter -ge $timeout ]; then
        echo "=== Go Server Failed to Start ===" >&2
        echo "Server failed to respond to health check after ${timeout} seconds" >&2
        exit 1
    fi
    counter=$((counter + 1))
    sleep 1
done

echo "Go server started successfully" >&2

# Start mock backend servers
echo "Starting mock backend servers..." >&2
cd /app && python3 start_backend_servers.py &
BACKEND_PID=$!

# Wait for backend servers to start (fixed time)
echo "Waiting for backend servers to start..." >&2
sleep 10

echo "Proceeding with tests..." >&2

# Run the tests and ensure we stay in the directory
cd /app
echo "Running tests..." >&2
python3 test_runner.py
TEST_EXIT_CODE=$?

# Wait a moment before checking for the file
sleep 2

# Verify test results file exists and has content
if [ -f "test_results.json" ] && [ -s "test_results.json" ]; then
    echo "Tests completed and results generated" >&2
    # Keep the processes running for a moment to ensure file is accessible
    sleep 2
    exit $TEST_EXIT_CODE
else
    echo "Failed to generate test results or file is empty" >&2
    # Print directory contents and file status for debugging
    echo "Directory contents:" >&2
    ls -la >&2
    exit 1
fi