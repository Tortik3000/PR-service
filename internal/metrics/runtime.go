package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

const namespace = "pr_service"

func init() {
	prometheus.Register(goCollector)
	prometheus.Register(processCollector)
}

var (
	processCollector = collectors.NewProcessCollector(
		collectors.ProcessCollectorOpts{},
	)

	goCollector = collectors.NewGoCollector(
		collectors.WithGoCollectorRuntimeMetrics(
			collectors.MetricsAll,
		),
	)
)
