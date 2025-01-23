package metrics

import "time"
import "github.com/prometheus/client_golang/prometheus"

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

func histogramObserve(histogramVec *prometheus.HistogramVec, value time.Duration, labels prometheus.Labels) error {
	histogram, err := histogramVec.GetMetricWith(labels)
	if err != nil {
		return err
	}
	histogram.Observe(float64(value.Microseconds()))
	return nil
}

const (
	LabelPeerId   = "peerId"
	LabelInfoHash = "infoHash"
	LabelReason   = "reason"
)

func registerCounter(registry *prometheus.Registry) map[counterMetrics]*prometheus.CounterVec {
	m := make(map[counterMetrics]*prometheus.CounterVec)

	m[CounterInvalidRequest] =
		prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: counterPrefix + string(CounterInvalidRequest),
			Help: "Total invalid announce counter",
		}, []string{LabelPeerId, LabelInfoHash, LabelReason})

	for _, h := range m {
		registry.MustRegister(h)
	}

	return m
}

func registerHistogram(registry *prometheus.Registry) map[string]*prometheus.HistogramVec {
	m := make(map[string]*prometheus.HistogramVec)

	for _, h := range m {
		registry.MustRegister(h)
	}

	return m
}
