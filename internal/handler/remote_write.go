package handler

import (
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
	"io"
	"log/slog"
	"net/http"
	"prom2click/internal/remotewrite"
)

func NewRemoteWriteHandler(batcher remotewrite.Batcher) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := slog.Default().With("handler", "RemoteWrite")

		// the healthcheck
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			logger.Error("Invalid method: ", r.Method)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("Could not read body from request: ", err)
			return
		}

		decompressed, err := snappy.Decode(nil, bodyBytes)

		var writeRequest prompb.WriteRequest
		if err := proto.Unmarshal(decompressed, &writeRequest); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("Could not unmarshal body: ", err)
			return
		}

		logger.Info("Received write request")
		err = batcher.Add(writeRequest)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("Could not add write request to batcher: ", err)
		}
	})
}
