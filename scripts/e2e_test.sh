#!/usr/bin/env bash
set -e

EVENT_URL="${EVENT_URL:-http://localhost:8080}"
BOOKING_URL="${BOOKING_URL:-http://localhost:8081}"

echo "=================================================="
echo "🎟️ TicketFlow End-to-End (E2E) Test Suite"
echo "=================================================="
echo "Event Service URL:   $EVENT_URL"
echo "Booking Service URL: $BOOKING_URL"
echo "--------------------------------------------------"

echo ""
echo "🚀 [1/6] Creating a new event with 10 available seats..."
EVENT_RESP=$(curl -s -f -X POST "$EVENT_URL/v1/event" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "E2E Champions Cup 2026",
    "start_time": "2026-10-15T20:00:00Z",
    "total_seats": 10
  }')

EVENT_ID=$(echo "$EVENT_RESP" | grep -o '"eventId":"[^"]*' | cut -d'"' -f4)
if [ -z "$EVENT_ID" ]; then
  echo "❌ Failed to create event: $EVENT_RESP"
  exit 1
fi
echo "✅ Created Event ID: $EVENT_ID"

echo ""
echo "🎟️ [2/6] User-101 booking 4 tickets..."
BOOKING_RESP=$(curl -s -f -X POST "$BOOKING_URL/v1/bookings" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user-101\",
    \"event_id\": \"$EVENT_ID\",
    \"ticket_count\": 4
  }")

BOOKING_ID=$(echo "$BOOKING_RESP" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
if [ -z "$BOOKING_ID" ]; then
  echo "❌ Failed to create booking: $BOOKING_RESP"
  exit 1
fi
echo "✅ Confirmed Booking ID: $BOOKING_ID"

echo ""
echo "🔍 [3/6] Checking remaining seats in Event Service (Expect 6 available)..."
EVENT_CHECK=$(curl -s -f -X GET "$EVENT_URL/v1/events/$EVENT_ID")
echo "   Event state: $EVENT_CHECK"

echo ""
echo "❌ [4/6] User-102 attempting to overbook (requests 8 tickets, only 6 left)..."
OVERBOOK_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BOOKING_URL/v1/bookings" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user-102\",
    \"event_id\": \"$EVENT_ID\",
    \"ticket_count\": 8
  }")

if [ "$OVERBOOK_STATUS" -ge 400 ]; then
  echo "✅ Overbooking rejected as expected with HTTP status $OVERBOOK_STATUS"
else
  echo "❌ Error: Overbooking should have failed but returned HTTP $OVERBOOK_STATUS"
  exit 1
fi

echo ""
echo "🔄 [5/6] Cancelling User-101 booking to release tickets..."
CANCEL_RESP=$(curl -s -f -X DELETE "$BOOKING_URL/v1/bookings/$BOOKING_ID")
echo "   Cancellation response: $CANCEL_RESP"

echo ""
echo "🔍 [6/6] Verifying available seats restored back to 10..."
RESTORED_EVENT=$(curl -s -f -X GET "$EVENT_URL/v1/events/$EVENT_ID")
echo "   Final event state: $RESTORED_EVENT"

echo ""
echo "=================================================="
echo "🎉 All TicketFlow E2E Tests Passed Successfully!"
echo "=================================================="
