.PHONY: all build test clean docs swagger run docker-build docker-run lint

# Set Go path
GO := /usr/local/go/bin/go

# Default target
all: lint test build

# Build the application
build:
	$(GO) build -o bin/api ./cmd/api

# Run tests
test:
	$(GO) test -v ./...

# Run linter // TODO: fix lint errors
lint:
	golangci-lint run --config .golangci.yml ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf docs/
	rm -rf data/

# Generate godoc documentation // TODO: fix docs errors
docs:
	@echo "Generating godoc documentation..."
	@mkdir -p docs/godoc
	godoc -all -html ./... > docs/godoc/index.html
	@echo "Documentation generated in docs/godoc/index.html"

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/api/main.go -o docs/swagger

# Run the application
run:
	$(GO) run cmd/api/main.go

# Build Docker image
docker-build:
	docker build -t gymshark-api .

# Run Docker container
docker-run:
	docker run -p 8080:8080 gymshark-api

# Install development dependencies
deps:
	# Install golangci-lint
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell $(GO) env GOPATH)/bin v1.55.2
	# Install other dependencies
	$(GO) install golang.org/x/tools/cmd/godoc@v0.19.0
	$(GO) install github.com/swaggo/swag/cmd/swag@latest
	$(GO) get github.com/swaggo/gin-swagger
	$(GO) get github.com/swaggo/files

# Help command
help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make test         - Run tests"
	@echo "  make lint         - Run linter - not available"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docs         - Generate godoc documentation - not available"
	@echo "  make swagger      - Generate Swagger documentation"
	@echo "  make run          - Run the application"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"
	@echo "  make deps         - Install development dependencies"
	@echo "  make help         - Show this help message" 