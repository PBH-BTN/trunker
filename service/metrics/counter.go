package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type counterMetrics string

const (
	metricsPrefix                         = "trunker_"
	CounterInvalidRequest  counterMetrics = "invalid_request_counter"
	CounterUDPRequest      counterMetrics = "udp_request_counter"
	CounterUDPRequestError counterMetrics = "udp_request_error_counter"
	CounterAnnounce        counterMetrics = "announce_counter"
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
	LabelAction = "action" // type of udp request
	LabelClient = "client"
	LabelSource = "source"
)

func registerCounter(registry *prometheus.Registry) map[counterMetrics]*prometheus.CounterVec {
	m := make(map[counterMetrics]*prometheus.CounterVec)

	m[CounterInvalidRequest] =
		prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: metricsPrefix + string(CounterInvalidRequest),
			Help: "Total invalid announce counter",
		}, []string{LabelReason})
	m[CounterUDPRequest] =
		prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: metricsPrefix + string(CounterUDPRequest),
			Help: "Total udp request counter",
		}, []string{LabelAction})
	m[CounterUDPRequestError] =
		prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: metricsPrefix + string(CounterUDPRequestError),
			Help: "Total udp request error counter",
		}, []string{LabelReason, LabelAction})
	m[CounterAnnounce] =
		prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: metricsPrefix + string(CounterAnnounce),
			Help: "Total announce counter",
		}, []string{LabelClient, LabelSource})
	for _, h := range m {
		registry.MustRegister(h)
	}

	return m
}
