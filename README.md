# Konnect Platform API

A RESTful API server for the Konnect Platform built with Go, Gin, PostgreSQL, and GORM. This service manages services and their versions for organizations.

## Overview
The server provides CRUD APIs for Konnect platform which contains organizations, their services and versions for each service
. It allows you to:

- **User Management**: Secure user registration and authentication with JWT tokens
- **Organization Management**: Create and manage organizations
- **Service Management**: Create and manage services with descriptions
- **Version Control**: Create and manage service versions
- **API Documentation**: Comprehensive Swagger/OpenAPI documentation
- **Testing**: Full integration test suite covering all endpoints

## Tech Stack

- **Language**: Go 1.24+
- **Web Framework**: Gin, used for the simple express style syntax which i am familiar with and find that it's easy to understand the flow
- **Database**: PostgreSQL, general purpose :)
- **ORM**: GORM (The application automatically runs database migrations on startup using GORM's AutoMigrate feature.), the ORM i am most familiar with, hence sticked with that, it's also widely used
- **Documentation**: Swagger/OpenAPI
- **Validation**: go-playground/validator (used as part of the gin framework itself, but have also used it separately before)
- **Logging**: https://github.com/uber-go/zap, fast and zero allocation logging

## Project Structure

```
.
├── controllers/          # HTTP handlers
│   ├── organization.go  # Organization endpoints
│   ├── service.go       # Service endpoints
│   ├── service_version.go # Service version endpoints
│   └── user.go          # User authentication endpoints
├── models/              # Database models and business logic
│   ├── base.go         # Base model with common fields
│   ├── organization.go # Organization model
│   ├── service.go      # Service model
│   ├── service_version.go # ServiceVersion model
│   └── user.go         # User model
├── pkg/                 # Reusable packages
│   ├── log/            # Structured logging with context
│   └── middleware/     # HTTP middlewares (auth, logging, CORS, etc.)
├── utils/               # Utility functions
│   ├── context.go      # Context helper functions
│   ├── jwt.go          # JWT token utilities
│   └── response.go     # HTTP response helpers
├── forms/               # Request validation structs
│   ├── organization.go # Organization form validation
│   ├── service.go      # Service form validation
│   ├── service_version.go # ServiceVersion form validation
│   ├── user.go         # User form validation
│   └── validator.go    # Custom validation rules (strong password)
├── routes/              # Route definitions
├── db/                  # Database configuration and migrations
├── tests/               # Integration tests
│   ├── helpers.go       # Test utilities and shared constants
│   ├── setup.go         # Test environment setup
│   ├── user_test.go     # User authentication tests
│   ├── organization_test.go # Organization API tests
│   ├── service_test.go  # Service API tests
│   └── service_version_test.go # Service version API tests
├── docs/                # Generated Swagger documentation
├── docker-compose.yml   # Docker setup
├── Makefile            # Build and development commands
└── main.go             # Application entry point
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
API_VERSION=1.0
DB_HOST=localhost:5432
DB_USER=your_db_user
DB_PASS=your_db_password
DB_NAME=konnect
JWT_SECRET=secret-key
TOKEN_CLEANUP_INTERVAL_MINUTES=60
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

## Testing

The project includes comprehensive integration tests covering all API endpoints.

### Running Tests

1. **Start test database**
```bash
docker-compose up -d
```
or start a postgres server and set `TEST_DB_HOST` env, default is `localhost:5433`

2. **Run all tests**
```bash
go test ./tests/ -v
```

3. **Run specific test**
```bash
go test ./tests/ -v -run TestUserRegistration
```

### Test Structure
- **User Tests**: Registration, login, logout with password validation
- **Organization Tests**: CRUD operations, access control
- **Service Tests**: Service management, query parameters, pagination
- **Service Version Tests**: Version creation, semantic versioning validation

All tests use a separate test database and include proper cleanup between test runs.

## Documentation

Interactive API documentation is available at `http://localhost:9000/swagger/index.html` when the server is running.

### Generate Swagger Documentation
Use the make command to generate swagger documentation if there are any changes made to the API spec(swagger comments)
```bash
make generate_docs
```

## Assumptions
- Currently only the user who created the organization belongs to that org, there is no feature to invite/add more users.
    - hence authorization is very basic, the user who created the org can do all operations on the org, haven't considered different sets of permissions for users as of now

## Trade offs
- Chose GORM's auto migrate for handling table schemas, works for most cases, no overhead, but a fully featured migration tool might be required for some setups

## Some implementation details
1. Authentication
    - Auth is being handled by generating custom JWT tokens and token invalidation on logout is being handled by maintaining the logged out tokens in DB and deleting them after expiry using a go routine which does the cleanup
    - Did not want to introduce redis(additional infra) for the token invalidation on logout use case, so went with postgres and a go routine which keeps cleaning up the table in background
2. Logs
    - JSON logs as they are easy to parse and transform outside of the application

## Improvements
1. Unit tests - repo currently only has integration tests for APIs as it covers most of the functionality
2. User invite flow - API to be able to invite/add user(s) to an org
    - invite links so that users can set their own password
    - an incremental functionality on top of this would be separate set of permissions for users
3. Add more logs
    - application supports log levels but currently only error logs are written in code, should also contain info and debug logs for improving logging
    - to make the most robust use of log levels, support for changing the log level run time should be added