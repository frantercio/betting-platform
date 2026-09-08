package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bets/bepping-platform/lib/auth"
	"github.com/bets/bepping-platform/services/wallet/internal/handler"
	"github.com/bets/bepping-platform/services/wallet/internal/ledger"
)

func main() {
	ctx := context.Background()
	var store ledger.Store
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		pg, err := ledger.NewPostgres(ctx, dsn)
		if err != nil {
			log.Fatalf("postgres init: %v", err)
		}
		store = pg
		log.Println("using postgres store")
	} else {
		store = ledger.NewMemStore()
		log.Println("using in-memory store")
	}
	svc := handler.New(store)

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
	}
	authMgr := auth.New([]byte(secret), 15*time.Minute)

	mux := http.NewServeMux()
	mux.Handle("GET /v1/wallet/balance", authMgr.Middleware(http.HandlerFunc(svc.Balance)))
	mux.Handle("POST /v1/wallet/deposits", authMgr.Middleware(http.HandlerFunc(svc.Deposit)))
	mux.Handle("POST /v1/wallet/withdrawals", authMgr.Middleware(http.HandlerFunc(svc.Withdraw)))
	mux.Handle("GET /v1/wallet/transactions", authMgr.Middleware(http.HandlerFunc(svc.Transactions)))
	mux.Handle("POST /v1/wallet/webhooks/pix", http.HandlerFunc(svc.PixWebhook))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := ":8080"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	log.Printf("wallet-service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}