# Build stage
FROM docker.io/library/golang:1.24-alpine@sha256:daae04ebad0c21149979cd8e9db38f565ecefd8547cf4a591240dc1972cf1399 AS builder

WORKDIR /app

# Copy source code
COPY . .

# Use vendored dependencies for build
RUN go build -mod=vendor -o /tmp/sign-blob-service ./cmd/server

# Final stage
FROM docker.io/library/alpine:latest@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1

WORKDIR /app

## Copy the binary from builder stage
COPY --from=builder /tmp/sign-blob-service .

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates && \
    # Create a non-root user
    addgroup -g 2222 appgroup && \
    adduser -D -s /bin/sh -u 2222 -G appgroup appuser && \
    mkdir -p /app/migrations && \
    chown -R appuser:appgroup /app && \
    chmod -R 774 /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 55555

# Command to run
## we are already in /app
CMD ["./sign-blob-service"]
