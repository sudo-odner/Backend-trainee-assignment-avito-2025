# Stage 1: Build
FROM golang:1.25.1 AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN mkdir -p /app/bin

RUN go build -o main ./cmd/main.go
CMD ["./main", "--config", "/app/config/docker.yaml"]