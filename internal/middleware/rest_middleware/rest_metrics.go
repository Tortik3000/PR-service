package rest_middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Tortik3000/PR-service/internal/metrics"
)

func MetricsMiddleware(service string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			next.ServeHTTP(ww, r)

			routePattern := chi.RouteContext(r.Context()).RoutePattern()
			metrics.RESTRequestDuration.
				WithLabelValues(service, routePattern).
				Observe(time.Since(start).Seconds())

			code := strconv.Itoa(ww.Status())

			metrics.RESTRequestsTotal.
				WithLabelValues(service, routePattern, code).
				Inc()

			if ww.Status() >= 200 && ww.Status() < 300 {
				metrics.RESTSuccessRequests.
					WithLabelValues(service, routePattern, code).
					Inc()
			}
		})
	}
}
