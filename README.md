# Smart Pack Allocation API

A Go-based API service that calculates optimal pack distribution for fulfilling orders with fixed pack sizes.

## Features

- RESTful API for pack allocation calculations
- Configurable pack sizes
- Optimal pack distribution algorithm
- Containerized deployment
- Comprehensive test coverage
- API documentation with Swagger

## Prerequisites

- Go 1.22 or later
- Docker and Docker Compose (for containerized deployment)
- Npm 9.5.1 or later
- Make (for development commands)

## Project Structure

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
├── frontend/        # Frontend
├── Makefile         # Development commands
└── README.md        # This file
```

## Getting Started

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

## API Usage

### Calculate Pack Distribution

```http
GET /calculate?quantity=500000
```

Example Response:

```json
{
    "packs": {
        "23": 37,
        "31": 29,
        "53": 9417
    },
    "total": 500000
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

## Documentation

### API Documentation (Swagger)

Generate and view the Swagger documentation:

```bash
make swagger
```

The Swagger UI will be available at <http://localhost:8080/swagger/index.html>

## Development Commands

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

## Testing

Run the test suite:

```bash
make test
```

## Configuration

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

## Frontend

The project includes a React-based frontend for interacting with the API.

### Running the Frontend Locally

1. Navigate to the frontend directory:

   ```bash
   cd frontend
   ```

2. Install dependencies:

   ```bash
   npm install
   ```

3. Start the development server:

   ```bash
   npm run dev
   ```

The frontend will be available at <http://localhost:3000>

### Note on Docker Support

Tried to run both in docker, but got some errors in the package.json, so decided I already spent too much time on the challenge anyway, so I moved my attention to the main features.

### Frontend Features

- Real-time pack allocation calculations
- Input validation
- Responsive design
- CORS support for local development
- Error handling and user feedback
