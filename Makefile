.PHONY: build run test clean

# Build the application
build:
	@echo "Building..."
	@go build -o bin/api cmd/api/main.go

# Run the application
run:
	@echo "Running..."
	@go run cmd/api/main.go

# Test the application
test:
	@echo "Testing..."
	@go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/api

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy

# All-in-one command to build and run
all: clean build run