package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	ProductsCacheHits = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "app",
		Subsystem: "cache",
		Name:      "products_hits_total",
		Help:      "Total number of product cache hits",
	})

	ProductsCacheMisses = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "app",
		Subsystem: "cache",
		Name:      "products_misses_total",
		Help:      "Total number of product cache misses",
	})
)

func Register() {
	prometheus.MustRegister(
		ProductsCacheHits,
		ProductsCacheMisses,
	)
}
