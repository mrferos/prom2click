# prom2click

## Notes

Setting up ClickHouse:
```bash
docker run -d --name clickhouse -p 18123:8123 -p19000:9000 --ulimit nofile=262144:262144 clickhouse/clickhouse-server
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