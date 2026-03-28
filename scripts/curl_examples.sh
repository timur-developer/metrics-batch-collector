#!/bin/sh

set -eu

BASE_URL="${BASE_URL:-http://localhost:8080}"

curl -i \
  -X POST "${BASE_URL}/events" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "page_view",
    "source": "landing",
    "user_id": "u123",
    "value": 1,
    "created_at": "2026-03-27T12:00:00Z"
  }'
