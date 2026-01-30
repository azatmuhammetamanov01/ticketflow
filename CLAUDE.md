# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Microservices-based Go application for managing event bookings. Uses gRPC for inter-service communication with PostgreSQL persistence. This is an educational project following a learning roadmap for microservices architecture.

**Tech Stack:** Go 1.25, gRPC, Protocol Buffers, PostgreSQL, gRPC-Gateway

## Build & Run Commands

### Booking Service
```bash
cd booking-service
go run ./server/main.go            # Runs gRPC on :50051 + HTTP gateway on :8080
go build -o booking-service ./server/main.go

# Regenerate proto code
protoc -I proto -I googleapis \
  --go_out=proto --go_opt=paths=source_relative \
  --go-grpc_out=proto --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=proto --grpc-gateway_opt=paths=source_relative \
  proto/booking.proto
```

### Event Service
```bash
cd event-service
go run ./server/main.go            # Runs gRPC on :50052
make proto                         # Regenerate proto code
make migration-create name=<name>  # Create new migration
make migration-up                  # Run migrations
```

### Dependencies
```bash
go mod tidy  # In each service directory
```

## Architecture

### Service Structure
Both services follow the same layered pattern:
```
*-service/
├── server/main.go              # Entry point
├── proto/*.proto               # gRPC service definitions
├── config/ or internal/config/ # Config loading from .env
├── internal/
│   ├── handler/grpc.go         # gRPC handlers (transport layer)
│   ├── service/                # Business logic layer
│   └── repository/             # PostgreSQL data layer
└── migrations/                 # SQL migration files
```

### Layer Responsibilities
1. **Handler** - gRPC request/response handling, error mapping to gRPC status codes
2. **Service** - Business logic, validation, custom errors (ErrNotFound, ErrInvalidInput, ErrInsufficientSeats)
3. **Repository** - PostgreSQL CRUD with parameterized queries

### Services
| Service | gRPC Port | HTTP Port | Database |
|---------|-----------|-----------|----------|
| booking-service | 50051 | 8080 | test_db |
| event-service | 50052 | - | test_db_2 |

## gRPC Endpoints

**Booking Service:**
- `CreateBooking` → POST `/v1/bookings`
- `GetBooking` → GET `/v1/bookings/{booking_id}`
- `ListUserBookings` → GET `/v1/users/{user_id}/bookings`
- `CancelBooking` → DELETE `/v1/bookings/{booking_id}`

**Event Service:**
- `CreateEvent` → POST `/v1/event`
- `GetEvent` → GET `/v1/events/{event_id}`
- `ListEvents` → GET `/v1/list/events`
- `UpdateAvailableTickets` → PUT `/v1/event/{event_id}`

## Configuration

Environment variables via `.env` files in each service:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`
- `HTTP_PORT`, `GRPC_PORT`, `SERVER_HOST`
- `APP_ENV` (development/production)

## Error Handling Pattern

```go
// Service layer defines errors
var ErrBookingNotFound = errors.New("booking not found")

// Handler maps to gRPC codes
if errors.Is(err, service.ErrBookingNotFound) {
    return nil, status.Error(codes.NotFound, err.Error())
}
```

## Package Imports

```go
// Booking Service
import pb "github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/proto"

// Event Service
import pb "github.com/azatmuhammetamanov01/online-ticket-booking/event-service/proto"
```

## Known Issues

- Event Service proto defines `EventSerive` (typo, missing 'c')
- Event Service has HTTP annotations in proto but no HTTP gateway implementation
- Inter-service gRPC communication not yet implemented
