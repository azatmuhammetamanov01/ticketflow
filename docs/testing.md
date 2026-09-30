# 🧪 TicketFlow Comprehensive Testing Guide

This document outlines the testing architecture, methodologies, execution strategies, and best practices for **TicketFlow**. It covers all levels of the testing pyramid: **Unit Tests**, **Integration Tests**, **End-to-End (E2E) Tests**, and **Stress / Concurrency Tests**.

---

## 📑 Table of Contents

1. [Testing Philosophy & Strategy](#-testing-philosophy--strategy)
2. [1. Unit Testing](#-1-unit-testing)
   - [How It Works](#how-unit-tests-work)
   - [Mocking Strategy](#mocking-strategy)
   - [Running Unit Tests](#running-unit-tests)
3. [2. Integration Testing](#-2-integration-testing)
   - [How It Works](#how-integration-tests-work)
   - [Repository & SQL Mocking](#repository--sql-mocking)
   - [Inter-Service gRPC Integration](#inter-service-grpc-integration)
4. [3. End-to-End (E2E) Testing](#-3-end-to-end-e2e-testing)
   - [How It Works](#how-e2e-tests-work)
   - [Full Lifecycle Test Scenario](#full-lifecycle-e2e-scenario)
   - [Automated E2E Test Script](#automated-e2e-test-script)
5. [4. Stress & Concurrency Testing](#-4-stress--concurrency-testing)
   - [How It Works & Flash Sale Simulation](#how-stress-tests-work)
   - [Preventing Race Conditions & Overselling](#preventing-overselling)
   - [Stress Test with k6](#stress-testing-with-k6)
   - [Stress Test with Go Goroutines](#stress-testing-with-go)
6. [📊 Testing Summary & Commands Cheat Sheet](#-testing-summary--commands-cheat-sheet)

---

## 🎯 Testing Philosophy & Strategy

```
               ▲
              / \
             /   \        Stress / Load Tests (Concurrency, Overselling, k6)
            /  ⚡ \
           /───────\      End-to-End (E2E) Tests (Full cross-service journeys)
          /  🌐 🌐  \
         /───────────\    Integration Tests (DB SQL queries, gRPC clients)
        /   📦   📦   \
       /───────────────\  Unit Tests (Domain, Usecases, Handlers with Mocks)
      /   🧪   🧪   🧪  \
     /───────────────────\
```

| Test Type | Scope | Dependencies | Speed | Primary Objective |
| :--- | :--- | :--- | :--- | :--- |
| **Unit** | Individual functions & structs | Mocked | Milliseconds | Validate pure business logic, input validation, and error paths |
| **Integration** | DB repositories & gRPC clients | `go-sqlmock` / Test DB | Seconds | Validate SQL correctness, schema mapping, and gRPC status translation |
| **End-to-End (E2E)** | Full multi-service workflow | All microservices & DBs | Seconds / Minutes | Verify end-user workflows across distributed boundaries |
| **Stress / Load** | Concurrency under peak traffic | Live staging / Docker cluster | Minutes | Detect race conditions, seat overselling, bottlenecks, and latency spikes |

---

## 🧪 1. Unit Testing

### How Unit Tests Work
Unit tests isolate a single component (such as a **Usecase** or a **gRPC Handler**) by replacing all external dependencies (databases, external gRPC clients) with in-memory **mocks**.

- **Pure Business Logic:** Verifies that [`BookingUsecase`](file:///home/azm/Documents/idea/ticketflow/booking-service/internal/usecase/booking.go) correctly validates inputs, checks available seats, calls ticket reservation, and creates records.
- **Error Propagation:** Tests edge cases such as `ErrInvalidInput`, `ErrEventNotFound`, `ErrInsufficientSeats`, and `ErrAlreadyCancelled`.

### Mocking Strategy
Both microservices use [`testify/mock`](https://github.com/stretchr/testify) to create mock implementations of domain interfaces:

```go
// Example: Mocking EventClient in Booking Service
eventClient := new(mocks.MockEventClient)
repo := new(mocks.MockBookingRepository)
uc := usecase.NewBookingUsecase(repo, eventClient)

// Define expectations
eventClient.On("GetEvent", ctx, "event-1").Return(&eventpb.Event{
    Id:             "event-1",
    AvailableSeats: 100,
}, nil)
eventClient.On("ReserveTickets", ctx, "event-1", int32(2)).Return(nil)
repo.On("Create", ctx, mock.AnythingOfType("*domain.Booking")).Return(nil)

// Execute and assert
booking, err := uc.CreateBooking(ctx, "user-1", "event-1", 2)
assert.NoError(t, err)
assert.Equal(t, "user-1", booking.UserID)
```

### Running Unit Tests

```bash
# Run all unit tests across the entire repository
go test ./...

# Run unit tests with verbose output and code coverage
go test -v -cover ./...

# Run unit tests with Go Race Detector
go test -race ./...

# Run tests for a specific service
cd booking-service && go test -v ./internal/usecase/...
cd event-service && go test -v ./internal/usecase/...
```

---

## 📦 2. Integration Testing

### How Integration Tests Work
Integration tests verify that components interact correctly with external systems such as **PostgreSQL** or other **gRPC services**.

### Repository & SQL Mocking
Repository tests in [`event-service/internal/repository/postgres/event_test.go`](file:///home/azm/Documents/idea/ticketflow/event-service/internal/repository/postgres/event_test.go) use [`go-sqlmock`](https://github.com/DATA-DOG/go-sqlmock) to simulate real database drivers without requiring a live PostgreSQL instance.

#### Example: Testing Atomic Seat Decrement
```go
func TestUpdateAvailableSeats_Success(t *testing.T) {
    db, mock, _ := sqlmock.New()
    repo := postgres.NewEventRepository(db)

    query := `
        UPDATE events
        SET available_seats = available_seats - $1
        WHERE id = $2 AND available_seats >= $1
        RETURNING available_seats
    `

    mock.ExpectQuery(regexp.QuoteMeta(query)).
        WithArgs(int32(2), "event-1").
        WillReturnRows(sqlmock.NewRows([]string{"available_seats"}).AddRow(int32(48)))

    newAvailable, err := repo.UpdateAvailableSeats(context.Background(), "event-1", 2)

    require.NoError(t, err)
    assert.Equal(t, int32(48), newAvailable)
}
```

### Inter-Service gRPC Integration
Integration tests for [`client.EventClient`](file:///home/azm/Documents/idea/ticketflow/booking-service/internal/client/event_client.go) verify:
- Accurate conversion between gRPC error codes (`codes.NotFound`, `codes.FailedPrecondition`) and domain errors (`ErrEventNotFound`, `ErrInsufficientSeats`).
- Proper connection lifecycle (`Close()`) and connection timeout handling.

---

## 🌐 3. End-to-End (E2E) Testing

### How E2E Tests Work
E2E tests treat the entire TicketFlow system as a black box. They start both microservices with live PostgreSQL databases and execute end-to-end user journeys over HTTP or gRPC.

### Full Lifecycle E2E Scenario

```
1. [POST /v1/event]             Create event with 10 available seats
           ↓
2. [POST /v1/bookings]          User A books 4 tickets (remaining: 6)
           ↓
3. [GET /v1/events/{id}]        Verify available seats equals 6
           ↓
4. [POST /v1/bookings]          User B attempts to book 8 tickets -> 409 Conflict (insufficient)
           ↓
5. [DELETE /v1/bookings/{id}]   User A cancels booking -> releases 4 seats (remaining: 10)
           ↓
6. [GET /v1/events/{id}]        Verify available seats restored to 10
```

### Automated E2E Test Script

Save this script as `scripts/e2e_test.sh` and execute against running services:

```bash
#!/usr/bin/env bash
set -e

EVENT_URL="http://localhost:8080"
BOOKING_URL="http://localhost:8081"

echo "🚀 [1/6] Creating a new event with 10 seats..."
EVENT_RESP=$(curl -s -X POST "$EVENT_URL/v1/event" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "E2E Champions Final",
    "start_time": "2026-10-15T20:00:00Z",
    "total_seats": 10
  }')

EVENT_ID=$(echo "$EVENT_RESP" | grep -o '"eventId":"[^"]*' | cut -d'"' -f4)
echo "✅ Created Event ID: $EVENT_ID"

echo "🎟️ [2/6] User-101 booking 4 tickets..."
BOOKING_RESP=$(curl -s -X POST "$BOOKING_URL/v1/bookings" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user-101\",
    \"event_id\": \"$EVENT_ID\",
    \"ticket_count\": 4
  }")

BOOKING_ID=$(echo "$BOOKING_RESP" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
echo "✅ Confirmed Booking ID: $BOOKING_ID"

echo "🔍 [3/6] Checking remaining seats in Event Service..."
EVENT_CHECK=$(curl -s -X GET "$EVENT_URL/v1/events/$EVENT_ID")
echo "Event state: $EVENT_CHECK"

echo "❌ [4/6] User-102 attempting to overbook (requests 8 tickets, only 6 left)..."
OVERBOOK_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BOOKING_URL/v1/bookings" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user-102\",
    \"event_id\": \"$EVENT_ID\",
    \"ticket_count\": 8
  }")

if [ "$OVERBOOK_STATUS" -ge 400 ]; then
  echo "✅ Overbooking properly rejected with HTTP $OVERBOOK_STATUS"
else
  echo "❌ Error: Overbooking was not rejected! Received HTTP $OVERBOOK_STATUS"
  exit 1
fi

echo "🔄 [5/6] Cancelling User-101 booking..."
CANCEL_RESP=$(curl -s -X DELETE "$BOOKING_URL/v1/bookings/$BOOKING_ID")
echo "Cancellation response: $CANCEL_RESP"

echo "🔍 [6/6] Verifying seats restored back to 10..."
RESTORED_EVENT=$(curl -s -X GET "$EVENT_URL/v1/events/$EVENT_ID")
echo "Final event state: $RESTORED_EVENT"

echo "🎉 All E2E Tests Passed Successfully!"
```

---

## ⚡ 4. Stress & Concurrency Testing

### How Stress Tests Work
Stress tests simulate a **Flash Sale / High Concurrency Spike** where hundreds of users attempt to purchase the same batch of limited seats in the exact same millisecond.

### Preventing Overselling
The system prevents overselling through atomic SQL execution:
```sql
UPDATE events
SET available_seats = available_seats - $1
WHERE id = $2 AND available_seats >= $1
RETURNING available_seats;
```
If 100 concurrent requests try to grab the last 5 seats, only the transactions that successfully execute `available_seats >= quantity` will commit; all others receive 0 rows updated and return `ErrInsufficientSeats`.

---

### Stress Testing with k6

Create a file `tests/stress/booking_stress.js`:

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    flash_sale: {
      executor: 'per-vu-iterations',
      vus: 50,              // 50 concurrent virtual users
      iterations: 2,        // 2 booking attempts per user (total 100 requests)
      maxDuration: '30s',
    },
  },
  thresholds: {
    http_req_failed: ['rate<1.0'], // We expect some 409s when sold out
    http_req_duration: ['p(95)<500'], // 95% of requests should finish under 500ms
  },
};

const EVENT_ID = __ENV.EVENT_ID || 'your-event-uuid-here';

export default function () {
  const url = 'http://localhost:8081/v1/bookings';
  const payload = JSON.stringify({
    user_id: `user-${__VU}-${__ITER}`,
    event_id: EVENT_ID,
    ticket_count: 1,
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  const res = http.post(url, payload, params);

  // Status is either 200/201 (Booked) or 409/400 (Sold out / Insufficient seats)
  check(res, {
    'valid response status': (r) => r.status === 200 || r.status === 201 || r.status === 409 || r.status === 400,
  });

  sleep(0.05);
}
```

#### Run k6:
```bash
k6 run -e EVENT_ID="<YOUR_EVENT_UUID>" tests/stress/booking_stress.js
```

---

### Stress Testing with Go

A native Go concurrency test sending 100 parallel booking requests against 10 tickets:

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
)

func main() {
	eventID := "YOUR_EVENT_ID"
	totalRequests := 100
	var successfulBookings int32
	var rejectedBookings int32

	var wg sync.WaitGroup
	wg.Add(totalRequests)

	fmt.Println("🚀 Starting concurrency test with 100 concurrent requests...")

	for i := 0; i < totalRequests; i++ {
		go func(userID int) {
			defer wg.Done()

			body, _ := json.Marshal(map[string]interface{}{
				"user_id":      fmt.Sprintf("user-%d", userID),
				"event_id":     eventID,
				"ticket_count": 1,
			})

			resp, err := http.Post("http://localhost:8081/v1/bookings", "application/json", bytes.NewBuffer(body))
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
				atomic.AddInt32(&successfulBookings, 1)
			} else {
				atomic.AddInt32(&rejectedBookings, 1)
			}
		}(i)
	}

	wg.Wait()

	fmt.Printf("✅ Successful Bookings: %d\n", successfulBookings)
	fmt.Printf("❌ Rejected / Sold Out: %d\n", rejectedBookings)
	fmt.Println("🔒 Zero overselling verified if Success == Event Total Seats!")
}
```

---

## 📊 Testing Summary & Commands Cheat Sheet

| Command | Purpose |
| :--- | :--- |
| `make test` | Run all unit tests across all Go workspace modules |
| `go test -v ./...` | Run tests with verbose output |
| `go test -race ./...` | Detect data races in concurrent code |
| `go test -coverprofile=cov.out ./... && go tool cover -html=cov.out` | Generate visual HTML test coverage report |
| `bash scripts/e2e_test.sh` | Run multi-service End-to-End workflow validation |
| `k6 run tests/stress/booking_stress.js` | Run high-concurrency load and overselling stress test |
| `grpcurl -plaintext localhost:9091 list` | Inspect live gRPC endpoints via reflection |
