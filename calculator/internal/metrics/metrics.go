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
	CalculationSuccessTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "calculation_success_total",
			Help: "Total number of successful cost calculations",
		},
		[]string{"method"},
	)

	CalculationFailureTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "calculation_failure_total",
			Help: "Total number of failed cost calculations",
		},
		[]string{"method", "reason"},
	)

	CalculatedCostValue = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "calculated_cost_value",
			Help: "Distribution of calculated delivery costs",
			Buckets: []float64{
				0,
				100,
				500,
				1000,
				5000,
				10000,
				20000,
				30000,
				40000,
				50000,
			},
		},
	)
)

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPResponseTime,
		CalculationSuccessTotal,
		CalculationFailureTotal,
		CalculatedCostValue,
	)
}
