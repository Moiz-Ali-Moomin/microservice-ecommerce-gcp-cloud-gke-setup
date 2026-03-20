package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter holds a per-IP token-bucket limiter.
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rps      rate.Limit
	burst    int
}

// NewRateLimiter creates an in-memory per-IP rate limiter.
//
//   - rps   : sustained requests per second per IP
//   - burst : maximum burst size (allows short spikes above rps)
//
// For production use at scale, replace the in-memory map with a Redis-backed
// implementation (e.g. github.com/go-redis/redis_rate).
func NewRateLimiter(rps int, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
}

// getLimiter returns (or lazily creates) a token-bucket limiter for the given key.
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if lim, ok := rl.limiters[key]; ok {
		return lim
	}

	lim := rate.NewLimiter(rl.rps, rl.burst)
	rl.limiters[key] = lim
	return lim
}

// RateLimit returns middleware that enforces the per-IP rate limit.
// Rejected requests receive a 429 Too Many Requests with Retry-After header.
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use X-Forwarded-For if behind a proxy/LB, fall back to RemoteAddr
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		}

		lim := rl.getLimiter(ip)
		if !lim.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Too many requests. Please retry after a short delay."}}`)) //nolint:errcheck
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequestID middleware injects a unique X-Request-ID header into every request
// and response, enabling end-to-end request tracing across services.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Propagate downstream and expose to caller
		r.Header.Set("X-Request-ID", requestID)
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

// generateRequestID creates a time-based unique ID.
// Uses timestamp + nanosecond precision for low-collision IDs without uuid dependency.
func generateRequestID() string {
	return "req-" + time.Now().Format("20060102150405.000000000")
}
