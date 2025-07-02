# Stage 1: Build the Go application
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to leverage Go module caching
COPY go.mod .
COPY go.sum .

# Download Go module dependencies
# This step is cached as long as go.mod/go.sum don't change
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
# -o main: specifies the output executable name as 'main'
# -ldflags "-s -w": reduces the binary size by stripping debug information
# . : builds the package in the current directory
RUN CGO_ENABLED=0 GOOS=linux go build -o main -ldflags "-s -w" .

# Stage 2: Create the final lean image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Expose the port your application listens on
# As per your main.go, it listens on port 8080
EXPOSE 8080

# Command to run the executable
# This is the entry point for your application when the container starts
CMD ["./main"]
