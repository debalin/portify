# Stage 1: Build the Go binary
FROM golang:1.25-alpine AS builder

# Add git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app statically
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Stage 2: Run the binary in a minimal container
FROM alpine:latest  

# Add CA certificates for HTTPS requests (Spotify/YouTube APIs require this)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/server .

# Expose port
EXPOSE 8080

# Command to run the executable
CMD ["./server"]
