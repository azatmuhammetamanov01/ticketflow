# 🧪 TicketFlow API Testing Guide

This guide details how to test both **REST (HTTP)** and **gRPC** endpoints across the **Event Service** and **Booking Service**.

---

## 🛠️ Prerequisites & Tools

- **cURL / Postman / Bruno** — For testing REST endpoints.
- **`grpcurl`** — For testing gRPC endpoints directly via server reflection.
  ```bash
  go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
  ```

---

## 🌐 1. REST (HTTP JSON) API Testing

Both services expose RESTful JSON APIs via **gRPC-Gateway**.

### A. Event Service (Default Port: `http://localhost:8080`, Docker Port: `8082`)

#### 1. Create a New Event (`POST /v1/event`)
```bash
curl -X POST http://localhost:8080/v1/event \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Rock Festival 2026",
    "start_time": "2026-09-01T20:00:00Z",
    "total_seats": 500
  }'
```
*Expected Response:*
```json
{
  "eventId": "a1b2c3d4-e5f6-7890-abcd-1234567890ab"
}
```

#### 2. List Events (`GET /v1/list/events`)
```bash
curl -X GET "http://localhost:8080/v1/list/events?limit=10&offset=0"
```

#### 3. Get Event Details (`GET /v1/events/{event_id}`)
```bash
curl -X GET http://localhost:8080/v1/events/<EVENT_ID>
```

#### 4. Update Available Tickets (`PUT /v1/event/{event_id}`)
```bash
curl -X PUT http://localhost:8080/v1/event/<EVENT_ID> \
  -H "Content-Type: application/json" \
  -d '{
    "quantity": 5
  }'
```

---

### B. Booking Service (Default Port: `http://localhost:8081`)

#### 1. Create a Booking (`POST /v1/bookings`)
> **Note:** The Booking Service verifies seat availability with the Event Service via gRPC before confirming.
```bash
curl -X POST http://localhost:8081/v1/bookings \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-101",
    "event_id": "<EVENT_ID>",
    "ticket_count": 2
  }'
```
*Expected Response (`201 Created`):*
```json
{
  "booking": {
    "id": "b8901234-...",
    "userId": "user-101",
    "eventId": "<EVENT_ID>",
    "ticketCount": 2,
    "status": "BOOKING_STATUS_PENDING",
    "createdAt": "2026-08-24T15:00:00Z"
  }
}
```

#### 2. Get Booking Details (`GET /v1/bookings/{booking_id}`)
```bash
curl -X GET http://localhost:8081/v1/bookings/<BOOKING_ID>
```

#### 3. List User Bookings (`GET /v1/users/{user_id}/bookings`)
```bash
curl -X GET http://localhost:8081/v1/users/user-101/bookings
```

#### 4. Cancel a Booking (`DELETE /v1/bookings/{booking_id}`)
```bash
curl -X DELETE http://localhost:8081/v1/bookings/<BOOKING_ID>
```

---

## ⚡ 2. gRPC API Testing (`grpcurl`)

Since **gRPC Server Reflection** is enabled on both microservices, you can discover and invoke gRPC methods directly.

### Discover Available Services
```bash
# Event Service (Port 9091)
grpcurl -plaintext localhost:9091 list

# Booking Service (Port 9091 or Docker mapping)
grpcurl -plaintext localhost:9091 list
```

### Event Service gRPC Calls

#### Create Event (`event.v1.EventService/CreateEvent`)
```bash
grpcurl -plaintext -d '{
  "name": "Tech Conference 2026",
  "start_time": "2026-11-01T09:00:00Z",
  "total_seats": 200
}' localhost:9091 event.v1.EventService/CreateEvent
```

#### Get Event (`event.v1.EventService/GetEvent`)
```bash
grpcurl -plaintext -d '{
  "event_id": "<EVENT_ID>"
}' localhost:9091 event.v1.EventService/GetEvent
```

#### List Events (`event.v1.EventService/ListEvents`)
```bash
grpcurl -plaintext -d '{
  "limit": 10,
  "offset": 0
}' localhost:9091 event.v1.EventService/ListEvents
```

### Booking Service gRPC Calls

#### Create Booking (`booking.v1.BookingService/CreateBooking`)
```bash
grpcurl -plaintext -d '{
  "user_id": "user-200",
  "event_id": "<EVENT_ID>",
  "ticket_count": 1
}' localhost:9091 booking.v1.BookingService/CreateBooking
```

---

## 🚫 3. Testing Error Scenarios (Negative Testing)

Genel API test standartlarında sadece başarılı senaryolar (Happy Path) değil, hatalı durumların da test edilmesi kritik önem taşır. İşte sık karşılaşılan hata durumları ve test yöntemleri:

### A. REST Hata Senaryoları Örnekleri

#### 1. Olmayan Event'e Bilet Alma (404 Not Found)
Geçersiz bir `event_id` göndererek sistemin doğru hata dönüp dönmediğini test edebilirsiniz:
```bash
curl -i -X POST http://localhost:8081/v1/bookings \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-101",
    "event_id": "invalid-event-id",
    "ticket_count": 2
  }'
```
*Beklenen Yanıt:* (404 Not Found veya 400 Bad Request)

#### 2. Yetersiz Koltuk Kapasitesi (409 Conflict veya 400 Bad Request)
Bir etkinlikte kalan koltuktan daha fazlasını almaya çalışın:
```bash
curl -i -X POST http://localhost:8081/v1/bookings \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-101",
    "event_id": "<EVENT_ID>",
    "ticket_count": 9999
  }'
```

#### 3. Eksik Veri ile İstek Atma (400 Bad Request)
Gerekli bir alanı (`name` gibi) göndermeyerek doğrulama (validation) kurallarını test edin:
```bash
curl -i -X POST http://localhost:8080/v1/event \
  -H "Content-Type: application/json" \
  -d '{
    "start_time": "2026-09-01T20:00:00Z"
  }'
```

### B. gRPC Hata Senaryoları Örnekleri

gRPC tarafında HTTP statüs kodları yerine gRPC durum kodları (gRPC Status Codes) döner (Örn: `NOT_FOUND`, `INVALID_ARGUMENT`).

#### Geçersiz Event ID ile Etkinlik Çağırma (`INVALID_ARGUMENT` veya `NOT_FOUND`)
```bash
grpcurl -plaintext -d '{
  "event_id": "non-existent-id"
}' localhost:9091 event.v1.EventService/GetEvent
```
*Beklenen Yanıt:*
```text
ERROR:
  Code: NotFound
  Message: event not found
```

---

## 🧪 3. Automated Go Tests (`go test`)

Run unit and mock integration tests across all microservices:

```bash
# Run all tests across the workspace via root Makefile
make test

# Run tests with verbose output for a specific service
cd event-service && go test ./... -v
cd booking-service && go test ./... -v
```
