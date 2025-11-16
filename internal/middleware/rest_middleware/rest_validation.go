package rest_middleware

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
)

func OpenAPIValidatorMiddleware(router routers.Router) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route, pathParams, err := router.FindRoute(r)
			if err != nil {
				http.Error(w, "Route not found", http.StatusNotFound)
				return
			}

			requestValidationInput := &openapi3filter.RequestValidationInput{
				Request:    r,
				PathParams: pathParams,
				Route:      route,
			}

			if err = openapi3filter.ValidateRequest(r.Context(), requestValidationInput); err != nil {
				http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
