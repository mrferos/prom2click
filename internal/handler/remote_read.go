package handler

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/prometheus/storage/remote"
	"log/slog"
	"net/http"
	"prom2click/internal/click"
)

func NewRemoteReadHandler(reader click.Reader) http.Handler {
	logger := slog.Default().With("handler", "RemoteRead")
	rrMetric := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "prom2click_remote_read_requests_total",
		Help: "The total number of remote read requests",
	}, []string{"status"})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr, err := remote.DecodeReadRequest(r)
		if err != nil {
			rrMetric.WithLabelValues("cannot_decode_read_request").Inc()
			w.WriteHeader(http.StatusBadRequest)
			logger.Error("Could not decode read request: ", err)
			return
		}

		qr, err := reader.Read(r.Context(), rr)
		if err != nil {
			rrMetric.WithLabelValues("cannot_read_from_clickhouse").Inc()
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("Could not read from clickhouse: ", err)
			return
		}

		//w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/x-protobuf")
		w.Header().Add("Content-Encoding", "snappy")
		if err := remote.EncodeReadResponse(qr, w); err != nil {
			rrMetric.WithLabelValues("cannot_encode_read_response").Inc()
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("Could not encode read response: ", err)
			return
		}

		logger.Info("Processed read request")
	})
}
