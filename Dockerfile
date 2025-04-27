LABEL authors="ramabmtr"

# Build stage
FROM golang:1.24.2-alpine3.21 AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main .
RUN go build -o migrate ./cmd/migrate

# Final stage
FROM alpine:3.21

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Run the application
CMD ["./main"]