package middleware

import (
	"net/http"
	"sync"
	"time"
)

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline'")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

type bucket struct {
	tokens int
	last   time.Time
}

type RateLimiter struct {
	limit  int
	window time.Duration
	mu     sync.Mutex
	buckets map[string]*bucket
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:    limit,
		window:   window,
		buckets:  make(map[string]*bucket),
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		now := time.Now()
		rl.mu.Lock()
		b, ok := rl.buckets[ip]
		if !ok {
			b = &bucket{tokens: rl.limit, last: now}
			rl.buckets[ip] = b
		}
		elapsed := now.Sub(b.last)
		// refill
		if elapsed > rl.window {
			b.tokens = rl.limit
			b.last = now
		}
		if b.tokens <= 0 {
			rl.mu.Unlock()
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		b.tokens--
		rl.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
