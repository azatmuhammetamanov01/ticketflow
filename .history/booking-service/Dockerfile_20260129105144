# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

#Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o booking-service ./server/main.go

#Run stage
FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/booking-service .
EXPOSE 9090 8080

CMD ["./booking-service"]