package click

import (
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/tsdb/chunkenc"
)

type Series struct {
	//curr
}

func (s Series) Labels() labels.Labels {
	//TODO implement me
	panic("implement me")
}

func (s Series) Iterator(iterator chunkenc.Iterator) chunkenc.Iterator {
	//TODO implement me
	panic("implement me")
}
