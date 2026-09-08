package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bets/bepping-platform/lib/auth"
	"github.com/bets/bepping-platform/lib/security"
	"github.com/bets/bepping-platform/services/admin/internal/handler"
)

func adminGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := r.Context().Value("user")
		if uid != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	svc := handler.New()

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
	}
	authMgr := auth.New([]byte(secret), 15*time.Minute)

	// rate limiter 5 req/min per IP for login
	loginLimiter := security.NewRateLimiter(5, time.Minute)

	mux := http.NewServeMux()
	mux.Handle("POST /admin/login", loginLimiter.Middleware(http.HandlerFunc(svc.Login)))
	mux.Handle("GET /admin/healthz", http.HandlerFunc(svc.Health))
	mux.Handle("POST /admin/logout", authMgr.Middleware(adminGuard(http.HandlerFunc(svc.Logout))))
	mux.Handle("GET /admin/users", authMgr.Middleware(adminGuard(http.HandlerFunc(svc.Users))))
	mux.Handle("GET /admin/wallets", authMgr.Middleware(adminGuard(http.HandlerFunc(svc.Wallets))))
	mux.Handle("GET /admin/bets", authMgr.Middleware(adminGuard(http.HandlerFunc(svc.Bets))))
	mux.Handle("GET /admin/audit", authMgr.Middleware(adminGuard(http.HandlerFunc(svc.Audit))))
	mux.Handle("GET /admin/metrics", authMgr.Middleware(adminGuard(http.HandlerFunc(svc.Metrics))))
	mux.Handle("POST /admin/users/block", authMgr.Middleware(adminGuard(http.HandlerFunc(svc.BlockUser))))
	mux.Handle("POST /admin/wallets/adjust", authMgr.Middleware(adminGuard(http.HandlerFunc(svc.AdjustBalance))))
	mux.Handle("/", http.FileServer(http.Dir("./web")))
	mux.Handle("GET /metrics", http.HandlerFunc(svc.Prometheus))

	handlerChain := security.SecurityHeaders(security.RequestID(security.Logger(mux)))

	addr := ":8083"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	log.Printf("admin-service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handlerChain))
}
