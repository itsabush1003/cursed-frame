package middleware

import (
	"net/http"

	"golang.org/x/time/rate"
)

type RateLimitMiddleware struct {
	limiter *rate.Limiter
}

func (rlm *RateLimitMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rlm.limiter.Allow() {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func NewRateLimitMiddleware(userNum int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: rate.NewLimiter(rate.Limit(userNum * 5), userNum * 20),
	}
}