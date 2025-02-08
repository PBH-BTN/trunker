package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type counterMetrics string

const (
	counterPrefix                        = "trunker_"
	CounterInvalidRequest counterMetrics = "invalid_request_counter"
)

func counterAdd(counterVec *prometheus.CounterVec, value int, labels prometheus.Labels) error {
	counter, err := counterVec.GetMetricWith(labels)
	if err != nil {
		return err
	}
	counter.Add(float64(value))
	return nil
}

func histogramObserve(histogramVec *prometheus.HistogramVec, value float64, labels prometheus.Labels) error {
	histogram, err := histogramVec.GetMetricWith(labels)
	if err != nil {
		return err
	}
	histogram.Observe(value)
	return nil
}

const (
	LabelReason = "reason"
)

func registerCounter(registry *prometheus.Registry) map[counterMetrics]*prometheus.CounterVec {
	m := make(map[counterMetrics]*prometheus.CounterVec)

	m[CounterInvalidRequest] =
		prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: counterPrefix + string(CounterInvalidRequest),
			Help: "Total invalid announce counter",
		}, []string{LabelReason})

	for _, h := range m {
		registry.MustRegister(h)
	}

	return m
}

func registerHistogram(registry *prometheus.Registry) map[counterMetrics]*prometheus.HistogramVec {
	m := make(map[counterMetrics]*prometheus.HistogramVec)

	for _, h := range m {
		registry.MustRegister(h)
	}

	return m
}
