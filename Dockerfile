# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-server ./cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/auth-server .

# Copy swagger assets
COPY --from=builder /app/api/swagger ./api/swagger

# Create keys directory
RUN mkdir -p keys

# Expose the multiplexed port
EXPOSE 8080

# Run the server
CMD ["./auth-server"]
