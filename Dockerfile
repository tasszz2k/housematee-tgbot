# Start from the official Go image for building the binary
FROM golang:1.22-alpine AS builder

# Set working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to enable dependency caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source
COPY . .

# Build the Go binary statically with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o housematee-tgbot ./cmd

# Use a minimal base image for runtime
FROM alpine:latest

# Set timezone and create necessary directories if needed
ENV TZ=Asia/Ho_Chi_Minh

# Install certificates and tzdata
RUN apk add --no-cache ca-certificates tzdata

# Set working directory in runtime container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/housematee-tgbot .

# Copy any necessary static/config files if needed
# COPY config.yaml .

# Expose the port if needed (optional, unless Fly config requires it)
EXPOSE 8080

# Run the application
CMD ["./housematee-tgbot"]