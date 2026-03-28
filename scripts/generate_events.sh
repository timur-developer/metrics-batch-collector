#!/bin/sh

set -eu

BASE_URL="${BASE_URL:-http://localhost:8080}"
COUNT="${COUNT:-50}"

i=1
while [ "$i" -le "$COUNT" ]; do
  payload=$(printf '{"event_type":"page_view","source":"generator","user_id":"u%03d","value":%d,"created_at":"2026-03-27T12:00:%02dZ"}' "$i" "$i" $((i % 60)))
  curl -sS \
    -X POST "${BASE_URL}/events" \
    -H "Content-Type: application/json" \
    -d "$payload" \
    > /dev/null
  i=$((i + 1))
done

printf 'Sent %s events to %s/events\n' "$COUNT" "$BASE_URL"
