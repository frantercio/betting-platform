package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bets/bepping-platform/lib/auth"
	"github.com/bets/bepping-platform/services/user/internal/handler"
	"github.com/bets/bepping-platform/services/user/internal/rg"
)

func main() {
	ctx := context.Background()
	var store rg.Store
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		pg, err := rg.NewPostgres(ctx, dsn)
		if err != nil {
			log.Fatalf("postgres init: %v", err)
		}
		store = pg
		log.Println("user-service using postgres store")
	} else {
		store = rg.NewMemStore()
		log.Println("user-service using in-memory store")
	}
	svc := handler.New(store)

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
	}
	authMgr := auth.New([]byte(secret), 15*time.Minute)

	mux := http.NewServeMux()
	mux.Handle("GET /v1/rg/limits", authMgr.Middleware(http.HandlerFunc(svc.GetLimits)))
	mux.Handle("PUT /v1/rg/limits", authMgr.Middleware(http.HandlerFunc(svc.SetLimit)))
	mux.Handle("POST /v1/rg/self-exclusion", authMgr.Middleware(http.HandlerFunc(svc.SelfExclude)))
	mux.Handle("POST /v1/rg/reality-check", authMgr.Middleware(http.HandlerFunc(svc.RealityCheck)))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := ":8081"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	log.Printf("user-service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}