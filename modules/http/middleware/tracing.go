package mware

import (
	"net/http"
)

func Tracing() MiddlewareOut {
	return MiddlewareOut{
		Mware: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Infof("Tracing called")
				next.ServeHTTP(w, r)
			})
		},
	}
}
