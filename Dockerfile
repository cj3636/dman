# Multi-stage Dockerfile for dman dotfile sync tool
# Build stage
FROM registry.tyss.io/ci-images/go1.24:latest AS builder

WORKDIR /src

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version injection
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X git.tyss.io/cj3636/dman/internal/buildinfo.Version=${VERSION} \
              -X git.tyss.io/cj3636/dman/internal/buildinfo.Commit=${COMMIT} \
              -X git.tyss.io/cj3636/dman/internal/buildinfo.BuildTime=${BUILD_TIME}" \
    -o dman ./cmd/dman

# Runtime stage
FROM alpine:3.19

# Install ca-certificates for HTTPS requests and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh -u 1000 dman

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /src/dman /usr/local/bin/dman

# Set executable permissions
RUN chmod +x /usr/local/bin/dman

# Create data directory with proper permissions
RUN mkdir -p /app/data && chown -R dman:dman /app

# Switch to non-root user
USER dman

# Expose default port
EXPOSE 3626

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD /usr/local/bin/dman status --json || exit 1

# Default command
ENTRYPOINT ["/usr/local/bin/dman"]
CMD ["serve", "--addr", ":3626"]