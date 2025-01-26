# Stage 1: Build the Go binary
FROM golang:1.23.4-alpine AS builder

# Set working directory
WORKDIR /app

# Copy necessary files
COPY go.mod go.sum ./
RUN go mod download

# Copy all project files
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /blog ./cmd/blog/main.go

# Stage 2: Create minimal production image
FROM alpine:latest

# Install CA certificates (needed for HTTPS)
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /blog ./

# Copy required directories
COPY ./config ./config
COPY ./cmd/blog/templates ./templates
COPY ./cmd/blog/content ./content

# Make binary executable
RUN chmod +x /app/blog

# Expose port (change 8080 to your app's port)
EXPOSE 8080

# Command to run the application
CMD ["./blog", "-config", "./config/config.toml"]
