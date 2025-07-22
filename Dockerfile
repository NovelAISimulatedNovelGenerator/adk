# Multi-stage build for ADK APIs
FROM golang:1.24-alpine AS builder

# Install build dependencies including gcc for CGO
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build plugins first (需要 CGO 支持)
RUN mkdir -p /app/plugins
#RUN CGO_ENABLED=1 GOOS=linux go build -buildmode=plugin -o /app/plugins/novel_flow_v4.so ./flows/novel_v4/main.go
#RUN CGO_ENABLED=1 GOOS=linux go build -buildmode=plugin -o /app/plugins/rag_test_flow.so ./flows/rag_test/main.go
RUN CGO_ENABLED=1 GOOS=linux go build -buildmode=plugin -o /app/plugins/test.so ./flows/test/main.go

# Build the apiserver binary (启用 CGO 以支持插件系统)
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-w -s' -o apiserver ./cmd/apiserver

# Final stage - minimal runtime
FROM alpine:latest

# Install ca-certificates and libc for CGO support
RUN apk --no-cache add ca-certificates tzdata libc6-compat

# Create non-root user
RUN addgroup -g 1001 appgroup && \
    adduser -D -s /bin/sh -u 1001 -G appgroup appuser

# Create app directory
WORKDIR /app

# Copy binary and plugins from builder
COPY --from=builder /app/apiserver .
COPY --from=builder /app/config.example.yaml ./config.example.yaml
COPY --from=builder /app/plugins ./plugins

# Set proper permissions
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set default environment variables
ENV ADK_CONFIG=/app/config.yaml
ENV ADDR=:8080

# Volume for plugins and config
VOLUME ["/app/plugins", "/app/config.yaml"]

# Run the binary
CMD ["./apiserver", "--config", "/app/config.yaml", "--addr", ":8080"]