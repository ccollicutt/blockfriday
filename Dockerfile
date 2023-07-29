# Use the official Golang image as the base image
FROM golang:1.20 AS build

# Set the working directory
WORKDIR /app

# Copy the Go mod and sum files to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code into the container
COPY main.go ./

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o admission-controller .

# Use a minimal base image for the final container
FROM alpine:latest

# Copy the binary from the previous build stage
COPY --from=build /app/admission-controller /admission-controller

# Set the entrypoint for the container
ENTRYPOINT ["/admission-controller"]
