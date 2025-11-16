package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func init() {
	err := prometheus.Register(DBQueryLatency)
	if err != nil {
		log.Warn("DBQueryLatency already register")
	}
}

var (
	DBQueryLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "db_query_latency_seconds",
			Help:      "Задержка запросов к БД по операциям",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)
