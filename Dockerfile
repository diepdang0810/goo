# Multi-stage Dockerfile with build argument for service selection
ARG SERVICE=app

FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the specified service
ARG SERVICE
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/${SERVICE} ./cmd/${SERVICE}

# Final stage
FROM alpine:latest

ARG SERVICE
ENV SERVICE_NAME=${SERVICE}

WORKDIR /root/

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/bin/${SERVICE} /app
COPY --from=builder /app/config /config
COPY --from=builder /app/migrations /migrations

# Run the service
CMD ["/app"]
