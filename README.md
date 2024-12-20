# prom2click

## Notes

Setting up dependencies:
```bash
docker-compose up -d 
```

Schema:
```clickhouse
CREATE TABLE metrics (
    timestamp DateTime,                -- Event time
    metric_name String,                -- Metric name or identifier
    labels Map(String, String),        -- Labels for dimensional data
    value Float64                      -- Metric value
)
ENGINE = MergeTree
PARTITION BY toStartOfHour(timestamp)       -- Partition by month for efficient queries
ORDER BY (metric_name, timestamp)      -- Order for fast range queries
SETTINGS index_granularity = 8192;    -- Adjust index granularity for performance
```

Starting vector:
```bash 
docker run -v $(pwd)/vector/:/etc/vector/ --rm timberio/vector:0.43.1-debian -w --config /etc/vector/vector.toml
```