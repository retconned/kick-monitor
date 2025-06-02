# ----------------------------
# 🏗️ Build Stage
# ----------------------------
	FROM golang:1.24-alpine AS builder

	# Install necessary packages
	RUN apk add --no-cache git ca-certificates

	# Set working directory
	WORKDIR /app

	# Copy Go module files
	COPY go.mod go.sum ./

	# Download dependencies
	RUN echo "📦 Downloading Go modules..." && \
			go mod download

	# Copy source code
	COPY . .

	# List files for verification
	RUN echo "📁 Verifying source files:" && \
			ls -la && \
			echo "🔍 Searching for main.go:" && \
			find . -name main.go

	# Build the application
	RUN echo "🔨 Building the Go application..." && \
			CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o kick-monitor cmd/kick-monitor/main.go

	# ----------------------------
	# 🚀 Final Stage
	# ----------------------------
	FROM alpine:latest

	# Install CA certificates
	RUN apk --no-cache add ca-certificates

	# Create a non-root user
	RUN addgroup -g 1001 kick-monitor && \
			adduser -D -s /bin/sh -u 1001 -G kick-monitor kick-monitor

	# Set working directory
	WORKDIR /home/kick-monitor

	# Copy the binary from the builder stage
	COPY --from=builder /app/kick-monitor .

	# Change ownership to the non-root user
	RUN chown -R kick-monitor:kick-monitor /home/kick-monitor

	# Switch to the non-root user
	USER kick-monitor

	# Expose port (adjust if necessary)
	EXPOSE 8080

	# Health check
	HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
			CMD pgrep kick-monitor || exit 1

	# Run the application
	CMD ["./kick-monitor"]

