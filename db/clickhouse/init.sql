CREATE DATABASE IF NOT EXISTS test_db;

CREATE TABLE test_db.logs (
    id UInt64,
    event_time DateTime,
    level String,
    message String
) ENGINE = MergeTree()
ORDER BY (id, event_time);
