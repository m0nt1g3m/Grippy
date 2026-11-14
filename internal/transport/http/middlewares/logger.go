package middlewares

import (
	"Grippy/internal/infrastructure/logging"
	"Grippy/internal/transport/http/response"
	"net/http"
)

func LoggingMiddleware(logger *logging.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			res := response.HTTPResponseHandler{ResponseWriter: w, Status: http.StatusOK, Size: 0}
			next.ServeHTTP(&res, r)
			logger.Infof("%s %s %d %s", r.Method, r.URL.Path, res.Status, http.StatusText(res.Status))
		})
	}
}
