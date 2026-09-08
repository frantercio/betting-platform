package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bets/bepping-platform/lib/auth"
	"github.com/bets/bepping-platform/services/betting/internal/bet"
	"github.com/bets/bepping-platform/services/betting/internal/handler"
)

func main() {
	ctx := context.Background()
	var store bet.Store
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		pg, err := bet.NewPostgres(ctx, dsn)
		if err != nil {
			log.Fatalf("postgres init: %v", err)
		}
		store = pg
		log.Println("betting-service using postgres store")
	} else {
		store = bet.NewMemStore()
		log.Println("betting-service using in-memory store")
	}

	// In prod these call wallet-service and user-service over HTTP/gRPC.
	wallet := &httpWallet{base: os.Getenv("WALLET_URL")}
	rg := &stubRG{maxStake: 1000000} // 10k BRL; connects to rg-service later
	live := func(eventID string) (float64, bool) { return 1.85, false }

	svc := handler.New(store, wallet, rg, live)

	authMgr := auth.New([]byte(os.Getenv("JWT_SECRET")), 15*time.Minute)

	mux := http.NewServeMux()
	// Go 1.22+ pattern routing
	h := authMgr.Middleware(mux)
	mux.HandleFunc("POST /v1/bets", svc.Place)
	mux.HandleFunc("GET /v1/bets", svc.List)
	mux.HandleFunc("POST /v1/bets/{id}/cashout", svc.Cashout)
	mux.HandleFunc("POST /v1/bets/{id}/settle", svc.Settle)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := ":8082"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	log.Printf("betting-service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, h))
}

// httpWallet is a thin forwarding client to the real wallet-service.
type httpWallet struct{ base string }

func (w *httpWallet) Debit(userID, requestID string, cents int64, ref string) error {
	return nil // TODO: POST {base}/v1/wallet/internal/debit
}
func (w *httpWallet) Credit(userID, requestID string, cents int64, ref string) error {
	return nil // TODO: POST {base}/v1/wallet/internal/credit
}

// stubRG allows max stake; replaced by rg-service CheckBet in production.
type stubRG struct{ maxStake int64 }

func (g *stubRG) CheckBet(userID string, stake int64) error {
	if stake > g.maxStake {
		return bet.ErrStakeTooLarge
	}
	return nil
}