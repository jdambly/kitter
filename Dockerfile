# Stage 1: Build the application
FROM golang:1.17 AS build

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o k8s-jitter ./cmd/k8s-jitter

# Stage 2: Package the binary
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the build stage
COPY --from=build /app/k8s-jitter .

# Expose port 8080 for the app
EXPOSE 8080

# Command to run the executable
#CMD ["./k8s-jitter"]
