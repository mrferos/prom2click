package click

import "github.com/prometheus/prometheus/storage"

type Queryable struct {
}

func (q Queryable) Querier(mint, maxt int64) (storage.Querier, error) {
	//TODO implement me
	panic("implement me")
}
