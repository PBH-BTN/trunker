package metrics

import (
	"sync"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/prometheus/client_golang/prometheus"
)

var registry *prometheus.Registry
var once sync.Once
var counterHandler map[counterMetrics]*prometheus.CounterVec
var histogramHandler map[counterMetrics]*prometheus.HistogramVec
var gaugeHandler map[counterMetrics]prometheus.Collector

func GetRegistry() *prometheus.Registry {
	if registry != nil {
		once.Do(Init)
	}
	return registry
}

func Init() {
	registry = prometheus.NewRegistry()
	if config.AppConfig.Tracker.EnableMetrics {
		counterHandler = registerCounter(registry)
		histogramHandler = registerHistogram(registry)
		gaugeHandler = registerGauge(registry)
	}
}

func EmitCounter(metrics counterMetrics, value int, labels prometheus.Labels) {
	if !config.AppConfig.Tracker.EnableMetrics {
		return
	}
	if handler, ok := counterHandler[metrics]; ok {
		_ = counterAdd(handler, value, labels)
	}
}
