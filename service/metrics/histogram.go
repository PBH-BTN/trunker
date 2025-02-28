package metrics

import "github.com/prometheus/client_golang/prometheus"

type histogramMetrics string

const (
	HistogramLatency histogramMetrics = "udp_request_latency_us"
)

func registerHistogram(registry *prometheus.Registry) map[histogramMetrics]*prometheus.HistogramVec {
	m := make(map[histogramMetrics]*prometheus.HistogramVec)
	m[HistogramLatency] = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    metricsPrefix + string(HistogramLatency),
			Help:    "Latency (microseconds) of UDP that had been application-level handled by the server.",
			Buckets: []float64{5000, 10000, 25000, 50000, 100000, 250000, 500000, 1000000},
		},
		[]string{LabelAction},
	)
	for _, h := range m {
		registry.MustRegister(h)
	}

	return m
}
