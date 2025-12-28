FROM golang:1.25-alpine3.21 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
# RUN go build -o orchestrator
RUN go build -o ./orchestrator ./cmd/orchestrator

CMD ["./orchestrator"]

