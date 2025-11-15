package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.Register(RESTRequestsTotal)
	prometheus.Register(RESTRequestDuration)
	prometheus.Register(RESTSuccessRequests)
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
