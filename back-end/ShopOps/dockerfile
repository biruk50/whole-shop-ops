FROM golang:1.25.3-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application
# Note: Adjust the path to your main.go file
RUN go build -o main ./Delivery/main.go

# Create a minimal runtime image
FROM alpine:latest

# Install necessary runtime dependencies if needed
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the built binary from builder stage
COPY --from=builder /app/main .
# Copy .env file if needed in runtime
COPY --from=builder /app/.env .env

# Expose port (adjust based on your application)
EXPOSE 8080

# TO RUN THE DOCKER IMAGE:
# docker run -d -p 8080:8080 --name fms-container biruk100/fms-app:latest

#docker pull biruk100
#docker run biruk100
#
# Local helper script: ./scripts/build-and-push.sh TAG  (defaults to 'latest')

# Run the application
CMD ["./main"]