package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// HTTP Metrics
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HttpResponseTimeSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_time_seconds",
			Help:    "Response time of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// WebSocket Metrics
	// auction handler func WebSocketStream
	WSConnectionsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ws_connections_total",
		Help: "Total number of WebSocket connections opened",
	})

	WSDisconnectionsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ws_disconnections_total",
		Help: "Total number of WebSocket disconnections",
	})

	WSMessagesReceivedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ws_messages_received_total",
		Help: "Total number of WebSocket messages received",
	})

	WSMessagesSentTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ws_messages_sent_total",
		Help: "Total number of WebSocket messages sent",
	})

	WSActiveConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ws_active_connections",
		Help: "Current number of active WebSocket connections",
	})

	WSConnectionDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "ws_connection_duration_seconds",
		Help:    "Duration of WebSocket connections",
		Buckets: prometheus.DefBuckets,
	})
)

func init() {
	prometheus.MustRegister(
		HttpRequestsTotal,
		HttpResponseTimeSeconds,
		WSConnectionsTotal,
		WSDisconnectionsTotal,
		WSMessagesReceivedTotal,
		WSMessagesSentTotal,
		WSActiveConnections,
		WSConnectionDuration,
	)
}
