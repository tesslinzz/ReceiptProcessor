#!/bin/bash

# Ensure Go is installed
if ! command -v go &> /dev/null
then
    echo "Go could not be found, please install Go first."
    exit 1
fi

# Run tests for the current package
echo "Running tests..."

# Run the Go tests
go test -v ./...

# Check if tests were successful
if [ $? -eq 0 ]; then
    echo "All tests passed successfully!"
else
    echo "Some tests failed. Please check the output above."
    exit 1
fi
