package remotewrite

import (
	"context"
	"errors"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/prometheus/prometheus/prompb"
	"prom2click/internal/click"
	"sync"
	"time"
)

type Batcher interface {
	Add(data prompb.WriteRequest) error
}

type BatcherImpl struct {
	dataCh         chan prompb.WriteRequest
	stopCh         chan struct{}
	maxBatchSize   int
	batchTimeout   time.Duration
	clickhouseConn driver.Conn
	wg             *sync.WaitGroup
}

func (b *BatcherImpl) Add(data prompb.WriteRequest) error {
	select {
	case b.dataCh <- data:
		return nil
	default:
		return errors.New("could not add data to batcher, channel is full")
	}
}

func (b *BatcherImpl) Stop() error {
	close(b.stopCh)
	err := b.clickhouseConn.Close()
	if err != nil {
		return err
	}
	b.wg.Wait()
	return nil
}

func (b *BatcherImpl) start() {
	ticker := time.NewTicker(b.batchTimeout)
	defer func() {
		ticker.Stop()
		b.wg.Done()
	}()

	batch, err := b.getBatch()
	if err != nil {
		panic(err)
	}

	seenSamples := 0

	for {
		select {
		case <-b.stopCh:
			return
		case data := <-b.dataCh:
			for _, ts := range data.Timeseries {
				var name string
				labels := make(map[string]string, len(ts.Labels))
				for _, label := range ts.Labels {
					labels[label.Name] = label.Value
				}

				if nsVal, ok := labels["__name__"]; ok {
					name = nsVal
					delete(labels, "__name__")
				}

				if name == "" {
					continue
				}

				for _, sample := range ts.Samples {
					batch.Append(sample.Timestamp, name, labels, sample.Value)
					seenSamples++

					if seenSamples >= b.maxBatchSize {
						batch.Send()
						batch, err = b.getBatch()
						if err != nil {
							panic(err)
						}
						seenSamples = 0
					}
				}
			}
		case <-ticker.C:
			batch.Send()
			batch, err = b.getBatch()
			if err != nil {
				panic(err)
			}
		}
	}
}

func (b *BatcherImpl) getBatch() (driver.Batch, error) {
	return b.clickhouseConn.PrepareBatch(context.Background(), "INSERT INTO metrics (timestamp, metric_name, labels, value) VALUES")
}

func NewBatcher() (*BatcherImpl, error) {
	maxBatchSize := 1000
	batchTimeout := 10 * time.Second
	wg := &sync.WaitGroup{}

	//TODO: make establishing clickhouse connection better
	conn, err := click.GetConnection()

	if err != nil {
		return nil, fmt.Errorf("could not connect to clickhouse: %w", err)
	}

	b := &BatcherImpl{
		wg:             wg,
		clickhouseConn: conn,
		dataCh:         make(chan prompb.WriteRequest, maxBatchSize*2),
		stopCh:         make(chan struct{}),
		maxBatchSize:   maxBatchSize,
		batchTimeout:   batchTimeout,
	}

	b.wg.Add(1)
	go b.start()

	return b, nil
}
