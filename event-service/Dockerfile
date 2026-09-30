# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY gen ./gen
COPY event-service ./event-service

WORKDIR /app/event-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o event-service ./cmd/server/main.go

# Run stage
FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/event-service/event-service .

EXPOSE 9091 8081

CMD ["./event-service"]