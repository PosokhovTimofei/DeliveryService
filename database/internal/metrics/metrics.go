package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	DeliveredPackagesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "delivered_packages_total",
			Help: "Total number of packages automatically marked as delivered",
		},
	)
	CreatedPackages = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "created_packages_total",
			Help: "Total number of packages created",
		},
	)

	UpdatedPackages = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "updated_packages_total",
			Help: "Total number of packages updated manually",
		},
	)

	FailedPackageCreations = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "failed_package_creations_total",
			Help: "Total number of failed package creation attempts",
		},
	)

	PackageDeliveryDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "package_delivery_duration_seconds",
			Help:    "Time taken from package creation to delivery",
			Buckets: prometheus.LinearBuckets(1, 2, 10),
		},
	)
)

func init() {
	prometheus.MustRegister(
		DeliveredPackagesTotal,
		CreatedPackages,
		UpdatedPackages,
		FailedPackageCreations,
		PackageDeliveryDuration,
	)
}
