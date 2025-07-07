package metrics

import (
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func IncrementActiveConnections() {
	ActiveConnection.Inc()
}
func DecreamentActiveConnections() {
	ActiveConnection.Dec()
}
func RecordMessageSent() {
	TotalMessagesSent.Inc()
}
func RecordRedisPubError() {
	RedisPublisherError.Inc()
}
func RecordRedisSubError() {
	RedisSubcriberError.Inc()
}
func ObserveMessageLatency(start time.Time) {
	duration := time.Since(start).Seconds()
	MessageLatency.Observe(duration)
}
func ObserveWsLatency(start time.Time, roomID string, msgType string) {
	wsDeliverylatency.WithLabelValues(roomID, msgType).Observe(float64(time.Since(start).Seconds()))
} //later change this type to Actual event type

func InstrumentHTTP(path string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rr := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rr, r)
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues(
			r.Method,
			path,
			http.StatusText(rr.statusCode),
		).Observe(duration)
	})
}
func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}
