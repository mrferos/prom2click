package handler

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/prometheus/storage/remote"
	"log/slog"
	"net/http"
	"prom2click/internal/remotewrite"
)

func NewRemoteWriteHandler(batcher remotewrite.Batcher) http.Handler {
	logger := slog.Default().With("handler", "RemoteWrite")
	rwMetric := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "prom2click_remote_write_requests_total",
		Help: "The total number of remote write requests",
	}, []string{"status"})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// the healthcheck
		if r.Method == http.MethodGet {
			rwMetric.WithLabelValues("ok_healthcheck").Inc()
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			rwMetric.WithLabelValues("invalid_method").Inc()
			w.WriteHeader(http.StatusMethodNotAllowed)
			logger.Error("Invalid method: ", r.Method)
			return
		}

		wr, err := remote.DecodeWriteRequest(r.Body)
		if err != nil {
			rwMetric.WithLabelValues("cannot_decode_write_request").Inc()
			w.WriteHeader(http.StatusBadRequest)
			logger.Error("Could not decode write request: ", err)
		}

		//logger.Info("Received write request")
		err = batcher.Add(*wr)
		if err != nil {
			rwMetric.WithLabelValues("cannot_add_to_batcher").Inc()
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("Could not add write request to batcher: ", err)
		} else {
			rwMetric.WithLabelValues("ok").Inc()
			w.WriteHeader(http.StatusOK)
		}
	})
}
