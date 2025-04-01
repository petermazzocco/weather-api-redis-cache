# Stage 1: Build the Go application
FROM --platform=$BUILDPLATFORM golang:1.23.2 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Set CGO_ENABLED=0 to create a statically linked binary
# Set GOOS and GOARCH for the target platform
ARG TARGETPLATFORM
RUN case "${TARGETPLATFORM}" in \
      "linux/amd64") GOARCH=amd64 ;; \
      "linux/arm64") GOARCH=arm64 ;; \
      *) GOARCH=amd64 ;; \
    esac && \
    CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build -o main .

# Stage 2: Create the final image
FROM --platform=$TARGETPLATFORM alpine:latest
WORKDIR /app

# Copy binary and application files from builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

# Install CA certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Make sure the binary is executable
RUN chmod +x ./main

# Create a non-root user to run the application
RUN adduser -D appuser && chown -R appuser:appuser /app
USER appuser

EXPOSE 8080
CMD ["./main"]