# 🎫 Online Ticket Booking System - Öğrenme Roadmap'i

Bu roadmap, microservices mimarisini öğrenmek için hazırlanmıştır. Her aşama bir öncekinin üzerine inşa edilir.

---

## 📋 Genel Bakış

### Öğreneceğin Teknolojiler
| Teknoloji | Ne İçin | Ne Zaman |
|-----------|---------|----------|
| gRPC + Protobuf | Servisler arası iletişim | Faz 1 |
| PostgreSQL | Veri depolama | Faz 1 |
| Docker | Containerization | Faz 2 |
| Kafka | Asenkron mesajlaşma | Faz 3 |
| Prometheus + Grafana | Monitoring | Faz 4 |
| Jaeger | Distributed tracing | Faz 4 |
| Kubernetes | Orchestration | Faz 5 (Opsiyonel) |

### Servis Mimarisi (Hedef)
```
┌─────────────────────────────────────────────────────────────┐
│                        API Gateway                           │
│                    (gRPC-Gateway / REST)                     │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│ Event Service │     │Booking Service│     │ User Service  │
│   (events)    │     │  (bookings)   │     │   (users)     │
└───────┬───────┘     └───────┬───────┘     └───────────────┘
        │                     │
        │         ┌───────────┴───────────┐
        │         ▼                       ▼
        │   ┌──────────┐           ┌──────────┐
        │   │  Kafka   │           │PostgreSQL│
        │   │ (events) │           │   (DB)   │
        │   └────┬─────┘           └──────────┘
        │        │
        │        ▼
        │   ┌──────────────┐
        │   │   Metrics    │
        │   │   Consumer   │
        │   └──────┬───────┘
        │          │
        ▼          ▼
┌─────────────────────────────────────────────────────────────┐
│              Prometheus → Grafana (Monitoring)               │
│                    Jaeger (Tracing)                          │
└─────────────────────────────────────────────────────────────┘
```

---

## 🚀 FAZ 1: gRPC ve Temel Booking Service (2-3 hafta)

### 1.1 gRPC Temelleri Öğren
**Hedef:** gRPC'nin ne olduğunu ve nasıl çalıştığını anla

**Öğrenilecekler:**
- [ ] Protocol Buffers (protobuf) nedir ve neden kullanılır
- [ ] gRPC vs REST farkları
- [ ] Unary RPC, Server streaming, Client streaming, Bidirectional streaming
- [ ] Proto dosyası yazımı

**Kaynaklar:**
- https://grpc.io/docs/languages/go/quickstart/
- https://protobuf.dev/programming-guides/proto3/

**Pratik:** `booking-service/proto/` altında basit bir proto dosyası yaz ve derle

### 1.2 Booking Service'i Tamamla
**Hedef:** Çalışan bir gRPC servisi oluştur

**Yapılacaklar:**

#### Adım 1: Proto Dosyasını Düzenle
```protobuf
// booking-service/proto/booking.proto
syntax = "proto3";

package booking;
option go_package = "./proto;proto";

import "google/protobuf/timestamp.proto";

service BookingService {
  rpc CreateBooking(CreateBookingRequest) returns (CreateBookingResponse);
  rpc GetBooking(GetBookingRequest) returns (GetBookingResponse);
  rpc ListUserBookings(ListUserBookingsRequest) returns (ListUserBookingsResponse);
  rpc CancelBooking(CancelBookingRequest) returns (CancelBookingResponse);
}

message Booking {
  string id = 1;
  string user_id = 2;
  string event_id = 3;
  int32 ticket_count = 4;
  string status = 5; // PENDING, CONFIRMED, CANCELLED
  google.protobuf.Timestamp created_at = 6;
}

message CreateBookingRequest {
  string user_id = 1;
  string event_id = 2;
  int32 ticket_count = 3;
}

message CreateBookingResponse {
  Booking booking = 1;
}

message GetBookingRequest {
  string booking_id = 1;
}

message GetBookingResponse {
  Booking booking = 1;
}

message ListUserBookingsRequest {
  string user_id = 1;
}

message ListUserBookingsResponse {
  repeated Booking bookings = 1;
}

message CancelBookingRequest {
  string booking_id = 1;
}

message CancelBookingResponse {
  bool success = 1;
  string message = 2;
}
```

#### Adım 2: Domain Modeli Oluştur
```
booking-service/internal/domain/booking.go
```

#### Adım 3: Repository Interface ve PostgreSQL Implementation
```
booking-service/internal/repository/repository.go (interface)
booking-service/internal/repository/postgres/booking.go (implementation)
```

#### Adım 4: Service Layer (Business Logic)
```
booking-service/internal/service/booking_service.go
```

#### Adım 5: gRPC Handler
```
booking-service/internal/transport/grpc/handler.go
```

#### Adım 6: Main Entry Point
```
booking-service/cmd/main.go
```

### 1.3 PostgreSQL Entegrasyonu
**Hedef:** Veritabanı bağlantısı ve CRUD operasyonları

**Yapılacaklar:**
- [ ] PostgreSQL'i local'de çalıştır (Docker ile)
- [ ] Database migration sistemi kur (golang-migrate)
- [ ] Connection pool ayarla
- [ ] Repository pattern ile CRUD yaz

**Docker ile PostgreSQL:**
```bash
docker run --name booking-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=1234 \
  -e POSTGRES_DB=booking_db \
  -p 5432:5432 \
  -d postgres:15
```

### 1.4 Test Yaz
**Hedef:** Birim ve entegrasyon testleri

- [ ] Service layer için unit testler
- [ ] Repository için integration testler (test container)
- [ ] gRPC handler için testler

---

## 🚀 FAZ 2: Event Service ve Servisler Arası İletişim (2 hafta)

### 2.1 Event Service Oluştur
**Hedef:** İkinci bir microservice yaz

**Event Service Sorumluluları:**
- Event (konser, maç, tiyatro) CRUD
- Koltuk/bilet kapasitesi yönetimi
- Tarih ve mekan bilgisi

**Proto Dosyası:**
```protobuf
// event-service/proto/event.proto
service EventService {
  rpc CreateEvent(CreateEventRequest) returns (CreateEventResponse);
  rpc GetEvent(GetEventRequest) returns (GetEventResponse);
  rpc ListEvents(ListEventsRequest) returns (ListEventsResponse);
  rpc UpdateAvailableTickets(UpdateTicketsRequest) returns (UpdateTicketsResponse);
}
```

### 2.2 Servisler Arası gRPC İletişimi
**Hedef:** Booking Service'in Event Service'i çağırması

**Senaryo:**
1. Kullanıcı booking yapmak ister
2. Booking Service → Event Service'e sorar: "Bu event var mı? Kapasite var mı?"
3. Event Service cevap verir
4. Booking Service booking'i oluşturur
5. Booking Service → Event Service'e söyler: "Kapasiteyi düşür"

**Öğrenilecekler:**
- [ ] gRPC client oluşturma
- [ ] Service discovery (şimdilik hardcoded, sonra Kubernetes DNS)
- [ ] Error handling ve retry logic
- [ ] Timeout ve deadline yönetimi

### 2.3 Docker Compose ile Local Development
**Hedef:** Tüm servisleri tek komutla ayağa kaldır

```yaml
# docker-compose.yml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: 1234
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  booking-service:
    build: ./booking-service
    ports:
      - "50051:50051"
    depends_on:
      - postgres
    environment:
      DB_HOST: postgres

  event-service:
    build: ./event-service
    ports:
      - "50052:50052"
    depends_on:
      - postgres

volumes:
  postgres_data:
```

---

## 🚀 FAZ 3: Kafka ve Asenkron İletişim (2 hafta)

### 3.1 Kafka Temelleri
**Hedef:** Event-driven architecture'ı anla

**Öğrenilecekler:**
- [ ] Kafka nedir, ne zaman kullanılır
- [ ] Topic, Partition, Consumer Group kavramları
- [ ] Producer ve Consumer yazımı
- [ ] Exactly-once vs At-least-once delivery

**Kaynaklar:**
- https://kafka.apache.org/documentation/
- https://github.com/segmentio/kafka-go (Go client)

### 3.2 Event Publishing
**Hedef:** Booking oluşturulduğunda event yayınla

**Senaryo:**
```
Booking Service                     Kafka                    Metrics Consumer
     │                                │                            │
     │ ──── BookingCreatedEvent ────► │                            │
     │                                │ ──── consume ────────────► │
     │                                │                            │
```

**Event Yapısı:**
```go
type BookingCreatedEvent struct {
    BookingID   string    `json:"booking_id"`
    UserID      string    `json:"user_id"`
    EventID     string    `json:"event_id"`
    TicketCount int       `json:"ticket_count"`
    CreatedAt   time.Time `json:"created_at"`
}
```

### 3.3 Metrics Consumer
**Hedef:** Kafka'dan event'leri consume et ve metrik topla

**metrics-consumer Görevi:**
- Booking event'lerini dinle
- İstatistik topla (günlük booking sayısı, popüler eventler, vb.)
- Prometheus'a metrik expose et

---

## 🚀 FAZ 4: Monitoring ve Observability (1-2 hafta)

### 4.1 Prometheus Metrikleri
**Hedef:** Servislerin metriklerini topla

**Eklenecek Metrikler:**
- `bookings_total` - Toplam booking sayısı
- `bookings_by_status` - Status'a göre booking (pending, confirmed, cancelled)
- `grpc_request_duration_seconds` - gRPC isteklerinin süresi
- `grpc_requests_total` - Toplam gRPC istek sayısı

**Go'da Prometheus:**
```go
import "github.com/prometheus/client_golang/prometheus"

var bookingsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "bookings_total",
        Help: "Total number of bookings",
    },
    []string{"status"},
)
```

### 4.2 Grafana Dashboard
**Hedef:** Metrikleri görselleştir

- [ ] Prometheus data source ekle
- [ ] Booking dashboard oluştur
- [ ] Alert kuralları yaz

### 4.3 Distributed Tracing (Jaeger)
**Hedef:** Request'leri servisler arasında takip et

**Öğrenilecekler:**
- [ ] OpenTelemetry SDK
- [ ] Trace, Span kavramları
- [ ] Context propagation
- [ ] Jaeger UI kullanımı

---

## 🚀 FAZ 5: Production Ready (Opsiyonel, 2+ hafta)

### 5.1 Kubernetes Deployment
- [ ] Deployment, Service, ConfigMap yazımı
- [ ] Health check (liveness, readiness probes)
- [ ] Resource limits
- [ ] Horizontal Pod Autoscaler

### 5.2 API Gateway
- [ ] gRPC-Gateway ile REST endpoint'ler
- [ ] Rate limiting
- [ ] Authentication/Authorization

### 5.3 CI/CD
- [ ] GitHub Actions ile test ve build
- [ ] Docker image push
- [ ] Kubernetes deploy

---

## 📁 Hedef Proje Yapısı

```
online-ticket-booking/
├── booking-service/
│   ├── cmd/
│   │   └── main.go                 # Entry point
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go           # Configuration
│   │   ├── domain/
│   │   │   └── booking.go          # Domain models
│   │   ├── repository/
│   │   │   ├── repository.go       # Interface
│   │   │   └── postgres/
│   │   │       └── booking.go      # PostgreSQL impl
│   │   ├── service/
│   │   │   └── booking.go          # Business logic
│   │   └── transport/
│   │       └── grpc/
│   │           └── handler.go      # gRPC handlers
│   ├── proto/
│   │   └── booking.proto
│   ├── migrations/
│   │   └── 001_create_bookings.up.sql
│   ├── Dockerfile
│   ├── go.mod
│   └── .env
│
├── event-service/
│   ├── cmd/
│   ├── internal/
│   ├── proto/
│   ├── migrations/
│   └── Dockerfile
│
├── metrics-consumer/
│   ├── cmd/
│   ├── internal/
│   │   ├── consumer/
│   │   └── metrics/
│   └── Dockerfile
│
├── docker-compose.yml              # Local development
├── docker-compose.prod.yml         # Production-like
│
├── prometheus/
│   └── prometheus.yml
│
├── grafana/
│   └── dashboards/
│
└── k8s/                            # Kubernetes manifests
    ├── booking-service/
    ├── event-service/
    └── infrastructure/
```

---

## ✅ Kontrol Listesi

### Faz 1 Tamamlandı mı?
- [ ] Proto dosyası yazıldı ve derlendi
- [ ] gRPC server çalışıyor
- [ ] PostgreSQL bağlantısı var
- [ ] CRUD operasyonları çalışıyor
- [ ] Testler yazıldı

### Faz 2 Tamamlandı mı?
- [ ] Event Service çalışıyor
- [ ] Servisler arası iletişim var
- [ ] Docker Compose ile her şey ayağa kalkıyor

### Faz 3 Tamamlandı mı?
- [ ] Kafka çalışıyor
- [ ] Event publishing yapılıyor
- [ ] Metrics Consumer event'leri tüketiyor

### Faz 4 Tamamlandı mı?
- [ ] Prometheus metrikleri toplanıyor
- [ ] Grafana dashboard var
- [ ] Jaeger tracing çalışıyor

---

## 🆘 Yardım İçin

Bu roadmap boyunca takıldığın her yerde bana sorabilirsin. Şunları yapabilirim:

1. **Kod yazma** - Her adım için gerçek kod yazabilirim
2. **Debug** - Hata aldığında çözüm bulabilirim
3. **Açıklama** - Kavramları detaylı açıklayabilirim
4. **Best practices** - Doğru yaklaşımları gösterebilirim

Başlamak için hazır olduğunda "Faz 1'e başlayalım" de!
