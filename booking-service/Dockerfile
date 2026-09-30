# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY gen ./gen
COPY booking-service ./booking-service

WORKDIR /app/booking-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o booking-service ./cmd/server/main.go

# Run stage
FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/booking-service/booking-service .

EXPOSE 9090 8080

CMD ["./booking-service"]