package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	once             sync.Once
	TotalActiveRooms = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_active_rooms",
		Help: "Number of total Active Rooms",
	})
	ActiveConnection = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "chat_active_connections",
		Help: "Number of active websocket connections",
	})
	TotalMessagesSent = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "total_chat_messages_sent",
		Help: "Total No of messages sent by the client",
	})
	RedisPublisherError = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chat_redis_publisher_error",
		Help: "Number of Redis Publisher Errors",
	})
	RedisSubcriberError = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chat_redis_subscriber_error",
		Help: "Number of redis Subscriber Errors",
	})
	MessageLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "chat_message_latency_seconds",
		Help: "Histogram of message handling latencies",
	})
	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_requestion_duration_miliseconds",
			Help:    "Histogram of response time for handler",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"Method", "Path", "Status"},
	)
	wsDeliverylatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "websocket_message_delivery_seconds",
			Help:    "Time Taken to deliver message to Websocket Client",
			Buckets: prometheus.ExponentialBuckets(0.0005, 2, 10),
		},
		[]string{"room_id", "message_type"},
	)
)

func Init() {
	once.Do(
		func() {

			prometheus.MustRegister(
				ActiveConnection,
				TotalMessagesSent,
				RedisPublisherError,
				RedisSubcriberError,
				MessageLatency,
				wsDeliverylatency,
				httpDuration,
				TotalActiveRooms,
			)
		})
}
