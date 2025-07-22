.PHONY: build run test clean docker-up docker-down

# Build the application
build:
	go build -o marketflow ./cmd/marketflow

# Run the application
run: build
	./marketflow

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f marketflow

# Start development environment
docker-up:
	docker-compose up -d

# Stop development environment
docker-down:
	docker-compose down

# Format code
fmt:
	gofumpt -w .

# Install dependencies
deps:
	go mod tidy
	go mod download
