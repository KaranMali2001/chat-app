package handler

import (
	"fmt"
	"net/http"

	"github.com/chat-app/internal"
	"github.com/chat-app/internal/config"
	"github.com/chat-app/internal/metrics"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {

	internal.SendJson(true, map[string]interface{}{
		"message": fmt.Sprintf("Health Route called in sever %s", config.AppConfig.Name),
	}, nil, w)
}

var HealthHandler = metrics.InstrumentHTTP("/health", http.HandlerFunc(HealthCheck))
