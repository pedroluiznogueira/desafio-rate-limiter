package httpmiddleware

import (
	"net"
	"net/http"

	"rate-limiter/internal/config"
	"rate-limiter/internal/limiter"
)

const blockedMessage = "you have reached the maximum number of requests or actions allowed within a certain time frame"

func RateLimit(cfg config.Config, lim *limiter.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("API_KEY")

			key := clientIP(r)
			limit := cfg.IPLimit

			if tokenLimit, ok := cfg.LimitFor(token); ok {
				key = token
				limit = tokenLimit
			}

			allowed, err := lim.Allow(r.Context(), key, limit)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(blockedMessage))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
