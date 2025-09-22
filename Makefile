## RUN APPLICATION
run:
	@echo -e "🚀 Running the application..."
	@go run main.go

## TEST COMMANDS
test:
	@echo -e "🧪 Running all tests..."
	@go test ./tests/... -v

test-coverage:
	@echo -e "📊 Running tests with coverage..."
	@go test ./tests/... -v -coverprofile=coverage.out -coverpkg=./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo -e "✅ Coverage report generated at coverage.html"

test-clean:
	@echo -e "🧹 Cleaning test artifacts..."
	@rm -f coverage.out coverage.html
	@echo -e "✅ Test artifacts cleaned!"

## INSTALL SWAG CLI TOOL & PACKAGES
install_swag:
	@echo -e "📥 Installing Swag CLI and dependencies..."
	@which swag >/dev/null 2>&1 || (echo -e "❌ Swag CLI not found! Installing now..." && go install github.com/swaggo/swag/cmd/swag@latest)
	@echo -e "🔄 Updating project dependencies for Swag..."
	@go mod tidy
	@go mod download
	@echo -e "✅ Swag installation complete!"

## GENERATE API DOCUMENTATION
generate_docs: install_swag
	@echo -e "📜 Generating API documentation using Swag..."
	@swag init
	@echo -e "✅ API documentation generated successfully!"

## BUILD APPLICATION
build:
	@echo -e "🔨 Building the application..."
	@go build -o bin/konnect-api *.go
	@echo -e "✅ Build complete! Binary available at bin/konnect-api"

## DOCKER COMMANDS
docker-build:
	@echo -e "🐳 Building Docker image..."
	@docker build -f docker/Dockerfile -t konnect-api:latest .
	@echo -e "✅ Docker image built successfully!"

docker-run:
	@echo -e "🚀 Running Docker container..."
	@docker run -d --name konnect-api -p 9000:9000 --env-file .env konnect-api:latest
	@echo -e "✅ Container started! API available at http://localhost:9000"

docker-stop:
	@echo -e "🛑 Stopping Docker container..."
	@docker stop konnect-api || true
	@docker rm konnect-api || true
	@echo -e "✅ Container stopped and removed!"

docker-logs:
	@echo -e "📋 Showing container logs..."
	@docker logs -f konnect-api