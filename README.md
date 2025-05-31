# 🧠 Smart Pack Allocation API

A Go-based API service that calculates optimal pack distribution for fulfilling orders with fixed pack sizes.

## 🚀 Features

- RESTful API for pack allocation calculations
- Configurable pack sizes
- Optimal pack distribution algorithm
- Containerized deployment
- Comprehensive test coverage
- API documentation with Swagger
- Detailed code documentation with godoc

## 🛠️ Prerequisites

- Go 1.22 or later
- Docker and Docker Compose (for containerized deployment)
- Make (for development commands)

## 🏗️ Project Structure

```
.
├── cmd/
│   └── api/           # Application entry point
├── internal/
│   ├── api/          # HTTP handlers
│   ├── allocator/    # Core business logic
│   └── storage/      # Persistence layer
├── docs/             # Generated documentation
├── data/             # SQLite database
├── config/           # Configuration files
├── Dockerfile        # Container definition
├── docker-compose.yml # Local development setup
├── Makefile         # Development commands
└── README.md        # This file
```

## 🚀 Getting Started

### Local Development

1. Clone the repository:

   ```bash
   git clone https://github.com/n-th/gymshark.git
   cd gymshark
   ```

2. Install dependencies:

   ```bash
   make deps
   ```

3. Run the application:

   ```bash
   make run
   ```

### Docker Deployment

1. Build and run using Docker Compose:

   ```bash
   docker-compose up --build
   ```

## 📝 API Usage

### Calculate Pack Distribution

```http
GET /calculate?quantity=500000
```

Example Response:

```json
{
    "packs": {
        "23": 2,
        "31": 7,
        "53": 9429
    }
}
```

### Get Recent Allocations

```http
GET /recent
```

Example Response:

```json
{
    "allocations": [
        {
            "ID": 1,
            "OrderQuantity": 2,
            "Packs": {
                "23": 1
            },
            "Total": 23,
            "CreatedAt": "2025-05-31T20:18:17Z"
        }
    ]
}
```

### Health Check

```http
GET /health
```

Example Response:

```json
{
    "status": "ok"
}
```

## 📚 Documentation

### API Documentation (Swagger)

Generate and view the Swagger documentation:

```bash
make swagger
```

The Swagger UI will be available at <http://localhost:8080/swagger/index.html>

## 🛠️ Development Commands

The project includes several Make commands to help with development:

```bash
make build        # Build the application
make test         # Run tests
make clean        # Clean build artifacts
make swagger      # Generate Swagger documentation
make run          # Run the application
make docker-build # Build Docker image
make docker-run   # Run Docker container
make deps         # Install development dependencies
make help         # Show all available commands
```

## 🧪 Testing

Run the test suite:

```bash
make test
```

## 📦 Configuration

Pack sizes can be configured in `config/config.yaml`:

```yaml
pack_sizes:
  - 23
  - 31
  - 53
```

## 🎯 Edge Cases

The service handles various edge cases:

- Zero quantity orders
- Orders smaller than the smallest pack size
- Large orders requiring multiple pack combinations
- Exact pack size matches

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.
