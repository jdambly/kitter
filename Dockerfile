# Stage 1: Build the application
FROM golang:latest AS build
WORKDIR /app
COPY ./ ./
# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kitter .

# Stage 2: Package the binary
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the build stage
COPY --from=build /app/kitter .
