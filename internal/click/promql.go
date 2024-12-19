package click

import (
	"context"
	"fmt"
	"github.com/prometheus/prometheus/promql"
	"time"
)

type WrappedPromQLEngine interface {
	NewInstantQuery(ctx context.Context, opts promql.QueryOpts, qs string, ts time.Time) (promql.Query, error)
	NewRangeQuery(ctx context.Context, opts promql.QueryOpts, qs string, start, end time.Time, interval time.Duration) (promql.Query, error)
}

type WrappedPromQLEngineImpl struct {
	engine    *promql.Engine
	queryable *Queryable
}

func (w *WrappedPromQLEngineImpl) NewInstantQuery(ctx context.Context, opts promql.QueryOpts, qs string, ts time.Time) (promql.Query, error) {
	return w.engine.NewInstantQuery(ctx, w.queryable, opts, qs, ts)
}

func (w *WrappedPromQLEngineImpl) NewRangeQuery(ctx context.Context, opts promql.QueryOpts, qs string, start, end time.Time, interval time.Duration) (promql.Query, error) {
	return w.engine.NewRangeQuery(ctx, w.queryable, opts, qs, start, end, interval)
}

func GetWrappedPromQLEngine() (WrappedPromQLEngine, error) {
	engine := promql.NewEngine(promql.EngineOpts{})
	conn, err := GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get clickhouse connection: %w", err)
	}

	return &WrappedPromQLEngineImpl{
		engine: engine,
		queryable: &Queryable{
			querier: &Querier{
				conn: conn,
			},
		},
	}, nil
}
