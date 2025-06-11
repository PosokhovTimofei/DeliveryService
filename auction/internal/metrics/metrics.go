package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	BidOpsCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bid_repository_operations_total",
			Help: "Total number of bid repository operations",
		},
		[]string{"method", "status"},
	)

	BidOpsDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bid_repository_duration_seconds",
			Help:    "Duration of bid repository operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	PackageOpsCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "package_repository_operations_total",
			Help: "Total number of package repository operations",
		},
		[]string{"method", "status"},
	)

	PackageOpsDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "package_repository_duration_seconds",
			Help:    "Duration of package repository operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)

var (
	AuctionStartedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "auctions_started_total",
			Help: "Total number of auctions started",
		},
	)
	AuctionFinishedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "auctions_finished_total",
			Help: "Total number of auctions finished",
		},
	)
	AuctionFinishedWithWinner = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "auctions_finished_with_winner_total",
			Help: "Number of auctions finished with winner",
		},
	)
	AuctionFinishedWithoutWinner = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "auctions_finished_without_winner_total",
			Help: "Number of auctions finished without winner",
		},
	)
	BidsPlacedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "bids_placed_total",
			Help: "Total number of bids placed",
		},
	)
	BidErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bid_errors_total",
			Help: "Count of bid placement errors",
		},
		[]string{"reason"},
	)
)

var (
	KafkaMessagesSent = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kafka_messages_sent_total",
		Help: "Total number of Kafka messages successfully sent",
	}, []string{"topic"})

	KafkaMessagesError = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kafka_messages_error_total",
		Help: "Number of Kafka message publishing errors",
	}, []string{"topic", "reason"})

	KafkaPublishDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kafka_publish_duration_seconds",
		Help:    "Kafka message publishing duration",
		Buckets: prometheus.DefBuckets,
	}, []string{"topic"})
)

var (
	KafkaConsumerRetries = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "kafka_consumer_retries_total",
		Help: "Total retries by Kafka consumer",
	})
	KafkaConsumerRestarts = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "kafka_consumer_restart_total",
		Help: "Kafka consumer restart attempts after failure",
	})
	KafkaConsumeDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "kafka_consume_duration_seconds",
		Help:    "Duration of kafka.Consume call",
		Buckets: prometheus.DefBuckets,
	})

	MessagesConsumed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "messages_consumed_total",
		Help: "Total number of consumed Kafka messages by topic",
	}, []string{"topic"})

	MessagesFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "messages_failed_total",
		Help: "Failed to handle Kafka message",
	}, []string{"topic", "reason"})

	DeliveryInitPublished = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "delivery_init_published_total",
		Help: "Delivery init events successfully published",
	})
)

func init() {
	prometheus.MustRegister(
		BidOpsCount,
		BidOpsDuration,
		PackageOpsCount,
		PackageOpsDuration,
		AuctionStartedTotal,
		AuctionFinishedTotal,
		AuctionFinishedWithWinner,
		AuctionFinishedWithoutWinner,
		BidsPlacedTotal,
		BidErrorsTotal,
		KafkaMessagesSent,
		KafkaMessagesError,
		KafkaPublishDuration,
		KafkaConsumerRetries,
		KafkaConsumerRestarts,
		KafkaConsumeDuration,
		MessagesConsumed,
		MessagesFailed,
		DeliveryInitPublished,
	)
}
