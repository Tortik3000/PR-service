package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func init() {
	err := prometheus.Register(RESTRequestsTotal)
	if err != nil {
		log.Warn("RESTRequestsTotal already register")
	}
	err = prometheus.Register(RESTRequestDuration)
	if err != nil {
		log.Warn("RESTRequestDuration already register")
	}
	err = prometheus.Register(RESTSuccessRequests)
	if err != nil {
		log.Warn("RESTSuccessRequests already register")
	}
}

var (
	RESTRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "rest",
		Name:      "requests_total",
		Help:      "Общее число rest-запросов",
	}, []string{"service", "method", "code"})

	RESTRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "rest",
		Name:      "request_duration_seconds",
		Help:      "Время обработки rest-запроса (в секундах)",
		Buckets:   prometheus.DefBuckets,
	}, []string{"service", "method"})

	RESTSuccessRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "rest",
		Name:      "requests_success_total",
		Help:      "Общее успешных число rest-запросов",
	}, []string{"service", "method", "code"})
)
