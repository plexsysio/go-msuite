package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

func New() *prometheus.Registry {
	r := prometheus.NewRegistry()

	r.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
			Namespace: "msuite",
		}),
		collectors.NewGoCollector(),
	)

	return r
}
