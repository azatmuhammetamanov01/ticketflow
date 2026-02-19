# 🎟️ TicketFlow

> Microservices-based ticket booking system built with Go, gRPC, and PostgreSQL

## 📋 Overview

TicketFlow is a backend system that handles online ticket bookings for events. It is built using a microservice architecture where services communicate via gRPC.

## 🏗️ Architecture

```
User → API Gateway → Booking Service ←→ Event Service
                          ↓                    ↓
                      Booking DB           Events DB
```

### Flow

1. User sends `POST /bookings` with `eventId` and `userId`
2. **Booking Service** receives the request and calls **Event Service** via gRPC to check seat availability
3. **Event Service** queries the Events DB for available seats and returns the result
4. If seats are available:
   - Booking Service saves the booking to Booking DB
   - Booking Service calls Event Service to reserve the seat (decrement seat count)
   - Returns `201 Created`
5. If no seats available:
   - Returns `409 Conflict` (sold out)

## 🛠️ Tech Stack

- **Language:** Go
- **Communication:** gRPC + gRPC-Gateway (REST)
- **Database:** PostgreSQL
- **Containerization:** Docker & Docker Compose

## 📦 Services

### Booking Service
- Handles booking creation and management
- Communicates with Event Service via gRPC to check and reserve seats
- Exposes REST API via gRPC-Gateway
- **Ports:** `9091` (gRPC), `8081` (HTTP)
- **Database:** `test_db_1`

### Event Service
- Manages events and seat availability
- Handles seat reservation and decrement logic
- **Ports:** `9092` (gRPC), `8082` (HTTP)
- **Database:** `test_db_2`

## 🚀 Getting Started

### Prerequisites

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Go 1.22+](https://golang.org/) (for local development)

### Run with Docker

**1. Create shared network:**
```bash
docker network create microservices-network
```

**2. Start Booking Service:**
```bash
cd booking-service
docker-compose up --build
```

**3. Start Event Service:**
```bash
cd event-service
docker-compose up --build
```

### Run Locally

```bash
# Booking Service
cd booking-service
go run cmd/main.go

# Event Service
cd event-service
go run cmd/main.go
```

## ⚙️ Environment Variables

Both services use the following environment variables (via `.env` file):

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `1234` |
| `DB_NAME` | Database name | `test_db_1` |
| `DB_SSLMODE` | SSL mode | `disable` |
| `HTTP_PORT` | HTTP server port | `8080` |
| `GRPC_PORT` | gRPC server port | `9091` |
| `SERVER_HOST` | Server host | `0.0.0.0` |
| `EVENT_SERVICE_ADDR` | Event service gRPC address | `event-service-event-service-1:9091` |

## 📡 API Endpoints

### Booking Service (`localhost:8081`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/bookings` | Create a new booking |
| `GET` | `/healthz` | Health check |

### Event Service (`localhost:8082`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/events` | List all events |
| `GET` | `/healthz` | Health check |

## 🗂️ Project Structure

```
ticketflow/
├── booking-service/
│   ├── cmd/
│   ├── internal/
│   ├── proto/
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── .env
├── event-service/
│   ├── cmd/
│   ├── internal/
│   ├── proto/
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── .env
└── README.md
```

## 🔒 Notes

- `.env` files are excluded from version control
- Services communicate via a shared Docker network (`microservices-network`)
- Each service has its own isolated PostgreSQL database
