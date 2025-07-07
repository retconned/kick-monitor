.PHONY: build run dev test clean docker-build docker-run docker-down db-create db-drop db-migrate

# Check if .env file exists and include it
ifneq (,$(wildcard .env))
    include .env
    export
endif

# --- Application Commands ---
build:
	@echo "Building Go application..."
	go build -o kick-monitor ./cmd/kick-monitor/main.go

dev:
	@echo "Running application..."
	DB_HOST=$(DEV_DB_HOST) DB_PORT=$(DEV_DB_PORT) DB_USER=$(DEV_DB_USER) DB_PASSWORD=$(DEV_DB_PASSWORD) DB_NAME=$(DEV_DB_NAME) PROXY_URL=$(DEV_PROXY_URL) JWT_SECRET=$(JWT_SECRET) go run cmd/kick-monitor/main.go

dev-air:
	@echo "Running application in development mode with hot reloading (using air)..."
	# Pass environment variables directly to air
	DB_HOST=$(DEV_DB_HOST) \
	DB_PORT=$(DEV_DB_PORT) \
	DB_USER=$(DEV_DB_USER) \
	DB_PASSWORD=$(DEV_DB_PASSWORD) \
	DB_NAME=$(DEV_DB_NAME) \
	PROXY_URL=$(DEV_PROXY_URL) \
	JWT_SECRET=$(JWT_SECRET) \
	air .

test:
	@echo "Running tests..."
	go test ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -f kick-monitor

# --- Docker Commands ---
docker-build:
	@echo "Building Docker image..."
	docker build -t kick-monitor .

docker-up: docker-build
	@echo "Running Docker containers (app and db)..."
	docker-compose up -d

docker-down:
	@echo "Stopping and removing Docker containers..."
	docker-compose down
