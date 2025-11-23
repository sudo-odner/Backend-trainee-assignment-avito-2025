FROM golang:1.25.1 AS builder

WORKDIR /app
COPY . .

RUN go mod download

# Линтер v2
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.6.2
ENV PATH="$PATH:$(go env GOPATH)/bin"
RUN golangci-lint run ./...


RUN go build -o main ./cmd/main.go

CMD ["./main", "--config", "/app/config/docker.yaml"]