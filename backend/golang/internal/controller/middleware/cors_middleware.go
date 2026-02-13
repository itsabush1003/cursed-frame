package middleware

import (
	"net/http"
)

type CorsMiddleware struct {
}

func (cm *CorsMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		next.ServeHTTP(w, r)
	})
}

func NewCorsMiddleware() *CorsMiddleware {
	return &CorsMiddleware{}
}
