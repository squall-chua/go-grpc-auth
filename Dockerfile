# Web UI build stage
FROM node:22-alpine AS web-builder

WORKDIR /app/web

COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ .
RUN npx nuxt generate

# Go build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Copy built web assets for go:embed
COPY --from=web-builder /app/web/.output/public ./web/.output/public

RUN CGO_ENABLED=0 GOOS=linux go build -o auth-server ./cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/auth-server .
COPY --from=builder /app/api/swagger ./api/swagger

RUN mkdir -p keys

EXPOSE 8080

CMD ["./auth-server"]
