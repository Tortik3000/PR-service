package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	log "github.com/sirupsen/logrus"
)

const namespace = "pr_service"

func init() {
	err := prometheus.Register(goCollector)
	if err != nil {
		log.Warn("goCollector already register")
	}
	err = prometheus.Register(processCollector)
	if err != nil {
		log.Warn("processCollector already register")
	}
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
