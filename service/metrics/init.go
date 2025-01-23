package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var registry *prometheus.Registry
var once sync.Once

func GetRegistry() *prometheus.Registry {
	if registry != nil {
		once.Do(Init)
	}
	return registry
}

func Init() {
	registry = prometheus.NewRegistry()
}
