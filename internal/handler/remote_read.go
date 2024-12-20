package handler

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/prometheus/storage/remote"
	"log/slog"
	"net/http"
)

func NewRemoteReadHandler() http.Handler {
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

		logger.Info("Received read request", rr)

	})
}
