package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_time_seconds",
			Help:    "Response time of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
	ValidateSuccessTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "validate_success_total",
			Help: "Total number of successful validate",
		},
		[]string{"method"},
	)

	ValidateFailureTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "validate_failure_total",
			Help: "Total number of failed validate",
		},
		[]string{"method", "reason"},
	)
)

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPResponseTime,
		ValidateFailureTotal,
		ValidateSuccessTotal,
	)
}
