# Event Service

Event Service is a Go backend that exposes event management APIs over both gRPC and HTTP.
The HTTP API is served through grpc-gateway, so the same gRPC methods are also available as REST endpoints.

## What It Does

- Creates events
- Fetches a single event by ID
- Lists events with pagination
- Updates available tickets for an event
- Exposes a health check endpoint at `/health-check`

## Requirements

- Go 1.25+
- PostgreSQL
- `protoc` and the Go protobuf plugins if you want to regenerate proto files
- `goose` if you want to run migrations with the `make migrate-up` target

## Configuration

The service reads configuration from environment variables or a `.env` file.

Default values:

```env
APP_ENV=development

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=product_db
DB_SSLMODE=disable

HTTP_PORT=8080
GRPC_PORT=9091
SERVER_HOST=0.0.0.0
```

If you run with Docker Compose, the service uses these values:

```env
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=1234
DB_NAME=test_db_2
DB_SSLMODE=disable
```

## Run Locally

1. Start PostgreSQL and create the database you want to use.
2. Apply the migration in `migrations/`.
3. Start the service.

Using `make`:

```bash
make migrate-up
make run
```

Or run directly:

```bash
go run ./cmd/server
```

The default ports are:

- HTTP: `8080`
- gRPC: `9091`

## Run With Docker Compose

The repository includes a PostgreSQL container and the event service.

```bash
docker compose up --build
```

Helpful commands:

```bash
make up
make down
make logs
```

When running with Compose, the ports are mapped to:

- HTTP: `http://localhost:8082`
- gRPC: `localhost:9092`
- PostgreSQL: `localhost:5434`

## Run Tests

Run the full test suite:

```bash
make test
```

Or:

```bash
go test ./...
```

If you want a faster focused check while working on app-level code, this repository already uses `go test ./internal/app` often for that slice.

## HTTP REST API

The HTTP API is exposed through grpc-gateway.

Base URL locally:

```text
http://localhost:8080
```

Base URL with Docker Compose:

```text
http://localhost:8082
```

### Health Check

```bash
curl http://localhost:8080/health-check
```

Response:

```json
{"status":"ok"}
```

### Create Event

```bash
curl -X POST http://localhost:8080/v1/event \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Tech Conference",
    "start_time": "2026-06-20T18:00:00Z",
    "total_seats": 100
  }'
```

### Get Event

```bash
curl http://localhost:8080/v1/events/<event_id>
```

### List Events

```bash
curl "http://localhost:8080/v1/list/events?limit=10&offset=0"
```

### Update Available Tickets

The API uses the quantity to update available seats.

```bash
curl -X PUT http://localhost:8080/v1/event/<event_id> \
  -H 'Content-Type: application/json' \
  -d '{
    "quantity": 2
  }'
```

## gRPC API

The gRPC server listens on `localhost:9091` by default.
Because server reflection is enabled, you can inspect and call methods with `grpcurl` without a local proto file.

### List Services

```bash
grpcurl -plaintext localhost:9091 list
```

### Describe the Service

```bash
grpcurl -plaintext localhost:9091 describe event.EventService
```

### Create Event

```bash
grpcurl -plaintext -d '{
  "name": "Tech Conference",
  "start_time": "2026-06-20T18:00:00Z",
  "total_seats": 100
}' localhost:9091 event.EventService/CreateEvent
```

### Get Event

```bash
grpcurl -plaintext -d '{
  "event_id": "<event_id>"
}' localhost:9091 event.EventService/GetEvent
```

### List Events

```bash
grpcurl -plaintext -d '{
  "limit": 10,
  "offset": 0
}' localhost:9091 event.EventService/ListEvents
```

### Update Available Tickets

```bash
grpcurl -plaintext -d '{
  "event_id": "<event_id>",
  "quantity": 2
}' localhost:9091 event.EventService/UpdateAvailableTickets
```

## API Summary

| gRPC Method | HTTP Route |
| --- | --- |
| `CreateEvent` | `POST /v1/event` |
| `GetEvent` | `GET /v1/events/{event_id}` |
| `ListEvents` | `GET /v1/list/events` |
| `UpdateAvailableTickets` | `PUT /v1/event/{event_id}` |

## Project Structure

- `cmd/server` - application entrypoint
- `internal/app` - database and server bootstrap
- `internal/config` - environment loading and defaults
- `internal/handler/grpc` - gRPC handlers
- `internal/usecase` - business logic
- `internal/repository/postgres` - PostgreSQL repository implementation
- `proto` - protobuf definitions and generated code
- `migrations` - database migrations

## Notes

- gRPC reflection is enabled, so `grpcurl` works out of the box.
- The HTTP layer is powered by grpc-gateway, so REST requests and gRPC calls hit the same service implementation.
- If the service fails to start, the most common cause is a missing database or incorrect `DB_*` environment values.