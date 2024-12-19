package click

import (
	"github.com/prometheus/prometheus/storage"
)

type Queryable struct {
	querier *Querier
}

func (q *Queryable) Querier(mint, maxt int64) (storage.Querier, error) {
	return q.querier, nil
}
