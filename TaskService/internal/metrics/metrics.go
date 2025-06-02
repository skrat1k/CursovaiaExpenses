package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	CreatedExpense = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "expense_created_total",
			Help: "Total number of created expense",
		},
	)

	GottenExpense = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "expense_gotten_total",
			Help: "Total number of gotten expense",
		},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "handler", "status"},
	)
)

func Register() {
	prometheus.MustRegister(CreatedExpense)
	prometheus.MustRegister(GottenExpense)

	prometheus.MustRegister(RequestDuration)
}
