package click

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/Masterminds/squirrel"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/annotations"
)

type Querier struct {
	conn driver.Conn
}

func (q *Querier) LabelValues(ctx context.Context, name string, hints *storage.LabelHints, matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	qb := squirrel.Select("DISTINCT arrayJoin(mapValues(labels)) AS label_value").From("metrics")
	qb = q.addMatchers(qb, matchers)
	if hints.Limit > 0 {
		qb = qb.Limit(uint64(hints.Limit))
	}

	sql, _, err := qb.ToSql()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build sql: %w", err)
	}

	rows, err := q.conn.Query(ctx, sql)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	results := []string{}
	for rows.Next() {
		var labelValue string
		if err := rows.Scan(&labelValue); err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, labelValue)
	}
	return results, nil, nil
}

func (q *Querier) LabelNames(ctx context.Context, hints *storage.LabelHints, matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	qb := squirrel.Select("DISTINCT arrayJoin(mapKeys(labels)) AS label_key").From("metrics")
	qb = q.addMatchers(qb, matchers)
	if hints.Limit > 0 {
		qb = qb.Limit(uint64(hints.Limit))
	}

	sql, _, err := qb.ToSql()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build sql: %w", err)
	}

	rows, err := q.conn.Query(ctx, sql)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	results := []string{}
	for rows.Next() {
		var labelKey string
		if err := rows.Scan(&labelKey); err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, labelKey)
	}
	return results, nil, nil
}

func (q *Querier) Close() error {
	if err := q.conn.Close(); err != nil {
		return fmt.Errorf(
			"failed to close clickhouse connection: %w",
			err)
	}
	return nil
}

func (q *Querier) Select(ctx context.Context, sortSeries bool, hints *storage.SelectHints, matchers ...*labels.Matcher) storage.SeriesSet {
	qb := squirrel.Select("*").From("metrics")
	qb = q.addMatchers(qb, matchers)

	if hints.Limit > 0 {
		qb = qb.Limit(uint64(hints.Limit))
	}

	if hints.Start != 0 {
		qb = qb.Where(squirrel.GtOrEq{"timestamp": hints.Start})
	}

	if hints.End != 0 {
		qb = qb.Where(squirrel.LtOrEq{"timestamp": hints.End})
	}

	sql, _, err := qb.ToSql()
	if err != nil {
		return &SeriesSet{}
	}

	rows, err := q.conn.Query(ctx, sql)
	if err != nil {
		return &SeriesSet{}
	}

	return &SeriesSet{rows: rows}
}

func (q *Querier) addMatchers(qb squirrel.SelectBuilder, matchers []*labels.Matcher) squirrel.SelectBuilder {
	for _, matcher := range matchers {
		switch matcher.Type {
		case labels.MatchEqual:
			qb = qb.Where(squirrel.Eq{matcher.Name: matcher.Value})
			break
		case labels.MatchNotEqual:
			qb = qb.Where(squirrel.NotEq{matcher.Name: matcher.Value})
			break
		case labels.MatchRegexp:
			qb = qb.Where(squirrel.Like{matcher.Name: matcher.Value})
			break
		case labels.MatchNotRegexp:
			qb = qb.Where(squirrel.NotLike{matcher.Name: matcher.Value})
			break
		}
	}

	return qb
}
