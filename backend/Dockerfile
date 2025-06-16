# Stage 1: Build Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copy everything inside backend, including migrations under pkg/
COPY . .

RUN go build -o main server.go

# Stage 2: Runtime image
FROM alpine:latest

WORKDIR /app

# Copy binary
COPY --from=builder /app/main .

# Copy everything needed including pkg (which has migrations)
COPY --from=builder /app/pkg ./pkg

EXPOSE 8080
CMD ["./main"]
