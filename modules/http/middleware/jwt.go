package mware

import (
	"net/http"
)

func JWT() MiddlewareOut {
	return MiddlewareOut{
		Mware: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Info("JWT called")
				next.ServeHTTP(w, r)
			})
		},
	}
}
