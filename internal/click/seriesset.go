package click

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/annotations"
)

type SeriesSet struct {
	rows driver.Rows
	err  error
}

func (s *SeriesSet) Next() bool {
	next := s.rows.Next()
	if !next {
		s.rows.Close()
	}

	return next
}

func (s *SeriesSet) At() storage.Series {

}

func (s *SeriesSet) Err() error {
	return s.rows.Err()
}

func (s *SeriesSet) Warnings() annotations.Annotations {
	return nil
}
