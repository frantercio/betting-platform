package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bets/bepping-platform/lib/auth"
	"github.com/bets/bepping-platform/services/wallet/internal/ledger"
)

func TestDepositIdempotency(t *testing.T) {
	store := ledger.NewMemStore()
	svc := New(store)
	userID := "user_1234"

	rec1 := callDeposit(svc, userID, "req-1", 100.00)
	if rec1.Code != http.StatusCreated {
		t.Fatalf("first deposit: got %d, want 201 (%s)", rec1.Code, strings.TrimSpace(rec1.Body.String()))
	}

	rec2 := callDeposit(svc, userID, "req-1", 100.00) // same request_id
	if rec2.Code != http.StatusOK && rec2.Code != http.StatusCreated {
		t.Fatalf("dup deposit: got %d, want 200/201", rec2.Code)
	}

	main, bonus, err := store.Balance(userID)
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	if main != 10000 {
		t.Fatalf("duplicate deposit credited balance twice: got %d cents, want 10000", main)
	}
	if bonus != 0 {
		t.Fatalf("bonus should be 0, got %d", bonus)
	}
}

func TestWithdrawInsufficientFunds(t *testing.T) {
	store := ledger.NewMemStore()
	svc := New(store)
	userID := "user_0001"
	callDeposit(svc, userID, "req-in-1", 50.00)

	rec := callWithdraw(svc, userID, "req-wd", 100.00)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("overdraft: got %d, want 422", rec.Code)
	}
}

func TestDepositThenBalance(t *testing.T) {
	store := ledger.NewMemStore()
	svc := New(store)
	userID := "user_3afe"
	callDeposit(svc, userID, "req-bal-1", 250.50)

	main, _, err := store.Balance(userID)
	if err != nil {
		t.Fatal(err)
	}
	if main != 25050 {
		t.Fatalf("balance mismatch: got %d cents want 25050", main)
	}
}

// helpers ---------------------------------------------------------------

func callDeposit(svc *Service, userID, reqID string, amount float64) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]any{
		"amount":     amount,
		"currency":   "BRL",
		"request_id": reqID,
		"pix_key":    userID + "@example.com",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/wallet/deposits", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	svc.Deposit(rec, auth.WithUser(req, userID))
	return rec
}

func callWithdraw(svc *Service, userID, reqID string, amount float64) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]any{
		"amount":     amount,
		"currency":   "BRL",
		"request_id": reqID,
		"pix_key":    userID + "@example.com",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/wallet/withdrawals", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	svc.Withdraw(rec, auth.WithUser(req, userID))
	return rec
}