package click

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/Masterminds/squirrel"
	"github.com/prometheus/prometheus/prompb"
	"time"
)

type Reader interface {
	Read(ctx context.Context, request *prompb.ReadRequest) (*prompb.ReadResponse, error)
}

func NewReader(conn driver.Conn) (Reader, error) {
	return &ReaderImpl{conn: conn}, nil
}

type ReaderImpl struct {
	conn driver.Conn
}

func (r *ReaderImpl) Read(ctx context.Context, request *prompb.ReadRequest) (*prompb.ReadResponse, error) {
	response := &prompb.ReadResponse{}
	for _, query := range request.Queries {
		qr, err := r.query(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to query: %w", err)
		}

		response.Results = append(response.Results, qr)
	}

	return response, nil
}

func (r *ReaderImpl) query(ctx context.Context, query *prompb.Query) (*prompb.QueryResult, error) {
	qr := &prompb.QueryResult{}
	qb := squirrel.Select(
		"metric_name",
		"labels",
		"timestamp",
		"value",
		"cityHash64(concat(metric_name, toString(labels))) AS hash").
		From("metrics")

	if query.StartTimestampMs != 0 {
		qb = qb.Where("timestamp >= toDateTime(? / 1000)", query.StartTimestampMs)
	}

	if query.EndTimestampMs != 0 {
		qb = qb.Where("timestamp <= toDateTime(? / 1000)", query.EndTimestampMs)
	}

	for _, label := range query.Matchers {
		if label.Name == "__name__" {
			qb = qb.Where(squirrel.Eq{"metric_name": label.Value})
			continue
		}

		switch label.Type {
		case prompb.LabelMatcher_EQ:
			qb = qb.Where("labels[?] = ?", label.Name, label.Value)
			break
		case prompb.LabelMatcher_NEQ:
			qb = qb.Where("NOT labels[?] = ?", label.Name, label.Value)
			break
		case prompb.LabelMatcher_RE:
			qb = qb.Where("labels[?] IS NOT NULL", label.Name)
			qb = qb.Where("extract(labels[?], ?) != ''", label.Name, label.Value)
			break
		case prompb.LabelMatcher_NRE:
			qb = qb.Where("labels[?] IS NOT NULL", label.Name)
			qb = qb.Where("NOT extract(labels[?], ?) != ''", label.Name, label.Value)
			break
		}
	}

	sql, params, err := qb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql: %w", err)
	}

	rows, err := r.conn.Query(ctx, sql, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	staging := map[uint64]stagingResults{}

	for rows.Next() {
		var metricName string
		var tsLabels map[string]string
		var timestamp time.Time
		var value float64
		var hash uint64

		if err := rows.Scan(&metricName, &tsLabels, &timestamp, &value, &hash); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if _, ok := staging[hash]; !ok {
			var ls []prompb.Label
			for k, v := range tsLabels {
				ls = append(ls, prompb.Label{
					Name:  k,
					Value: v,
				})
			}

			staging[hash] = stagingResults{
				labels:  ls,
				samples: []prompb.Sample{},
			}
		} else {
			sr := staging[hash]
			sr.samples = append(sr.samples, prompb.Sample{
				Timestamp: timestamp.UnixMilli(),
				Value:     value,
			})

			staging[hash] = sr
		}
	}

	for _, sr := range staging {
		qr.Timeseries = append(qr.Timeseries, &prompb.TimeSeries{
			Labels:  sr.labels,
			Samples: sr.samples,
		})
	}

	return qr, nil
}

type stagingResults struct {
	labels  []prompb.Label
	samples []prompb.Sample
}
