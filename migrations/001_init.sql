CREATE TABLE IF NOT EXISTS events
(
    event_type String,
    source String,
    user_id String,
    value Float64,
    created_at DateTime
)
ENGINE = MergeTree
ORDER BY (created_at, event_type, user_id);
