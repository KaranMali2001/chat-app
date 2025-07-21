package handler

import (
	"net/http"

	"github.com/chat-app/internal/metrics"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("SERVER IS RUNNING"))
}

var HealthHandler = metrics.InstrumentHTTP("/health", http.HandlerFunc(HealthCheck))
