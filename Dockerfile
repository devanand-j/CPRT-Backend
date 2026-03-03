# Start from the official Golang base image
FROM golang:1.24 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first to leverage caching
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app for production
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/api

# Final image using a minimal base
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/server .

# Copy migration files
COPY --from=builder /app/migrations ./migrations

# Expose the application port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/app/server"]