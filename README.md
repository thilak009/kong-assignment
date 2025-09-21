# Konnect Platform API

A RESTful API server for the Konnect Platform built with Go, Gin, PostgreSQL, and GORM. This service manages services and their versions.

## Overview
The server provides CRUD APIs for Konnect platform which contains services and versions for each service
. It allows you to:

- Create and manage services with descriptions
- Create and manage service versions
- Enforce unique constraints (service + version combinations)
- Comprehensive API documentation with Swagger

## Tech Stack

- **Language**: Go 1.24+
- **Web Framework**: Gin
- **Database**: PostgreSQL
- **ORM**: GORM (The application automatically runs database migrations on startup using GORM's AutoMigrate feature.)
- **Documentation**: Swagger/OpenAPI
- **Validation**: go-playground/validator

## Project Structure

```
.
├── controllers/          # HTTP handlers
│   ├── service.go       # Service endpoints
│   └── service_version.go # Service version endpoints
├── models/              # Database models
│   ├── base.go         # Base model with common fields
│   ├── service.go      # Service model
│   └── service_version.go # ServiceVersion model
├── forms/               # Request validation structs
│   ├── service.go      # Service form validation
│   └── service_version.go # ServiceVersion form validation
├── db/                  # Database configuration
├── docs/                # Generated Swagger documentation
├── docker-compose.yml   # Docker setup
├── Makefile            # Build and development commands
└── main.go             # Application entry point```
```

## Getting Started

### Prerequisites

- Go 1.24+
- PostgreSQL 13+
- Docker & Docker Compose (optional)

### Environment Setup

1. **Clone the repository**
```bash
git clone <repository-url>
cd kong-assignment
```

2. **Copy environment file**
```bash
cp .env.example .env
```

3. **Configure environment variables**
```env
ENV=LOCAL
PORT=9000
API_VERSION=2.0
DB_HOST=localhost:5432
DB_USER=your_db_user
DB_PASS=your_db_password
DB_NAME=konnect
```

### Running Locally

1. **Start the services**
```bash
docker-compose up -d
```

This will start:
- PostgreSQL database on port 5433
- The API server will connect to this database

If you are not using docker, make sure to start a postgres server before running the application

2. **Run the application**
```bash
go run main.go
```
This would download necessary packages(if not already) and start the server

The server will be available at `http://localhost:9000` (default is set to 9000 in env)

## Documentation

Interactive API documentation is available at `http://localhost:9000/swagger/index.html` when the server is running.

### Generate Swagger Documentation
Use the make command to generate swagger documentation if there are any changes made to the API spec(swagger comments)
```bash
make generate_docs
```


## TODO
1. Search, Sort, Pagination
2. Metadata in services APIs
3. Tests
4. Authentication
5. Improve logging