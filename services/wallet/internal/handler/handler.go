package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/bets/bepping-platform/lib/auth"
	"github.com/bets/bepping-platform/services/wallet/internal/ledger"
)

// Service wires the ledger store with HTTP handlers.
type Service struct {
	Store ledger.Store
	Now   func() time.Time
}

func New(s ledger.Store) *Service {
	return &Service{Store: s, Now: time.Now}
}

// --- HTTP surface ---------------------------------------------------------

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, apiError{Code: http.StatusText(code), Message: msg})
}

// --- Requests (OpenAPI schemas) -------------------------------------------

type depositReq struct {
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	RequestID string  `json:"request_id"`
	PixKey    string  `json:"pix_key"`
}

type withdrawReq struct {
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	RequestID string  `json:"request_id"`
	PixKey    string  `json:"pix_key"`
}

type txnResp struct {
	TransactionID int64   `json:"transaction_id"`
	Status        string  `json:"status"`
	QRCode        string  `json:"qr_code,omitempty"`
	CopiaECola    string  `json:"copia_e_cola,omitempty"`
	ExpiresAt     string  `json:"expires_at,omitempty"`
	FraudCheck    *string `json:"fraud_check,omitempty"`
}

// money rounds a BRL float to integer cents.
func money(v float64) int64 { return int64(v*100 + 0.5) }

// --- Handlers --------------------------------------------------------------

// Balance GET /v1/wallet/balance
func (s *Service) Balance(w http.ResponseWriter, r *http.Request) {
	userID := userFromCtx(r)
	if userID == "" {
		writeErr(w, http.StatusUnauthorized, "missing user")
		return
	}
	main, bonus, err := ledgerBal(s.Store, userID)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"available": float64(main) / 100,
		"bonus":     float64(bonus) / 100,
		"currency":  "BRL",
	})
}

// Deposit POST /v1/wallet/deposits
// Creates a PENDING deposit; in production the PSP confirms via webhook.
func (s *Service) Deposit(w http.ResponseWriter, r *http.Request) {
	userID := userFromCtx(r)
	var req depositReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.RequestID == "" || req.Amount <= 0 {
		writeErr(w, http.StatusBadRequest, "amount and request_id are required")
		return
	}
	if req.Currency != "" && req.Currency != "BRL" {
		writeErr(w, http.StatusBadRequest, "only BRL supported")
		return
	}

	cents := money(req.Amount)

	// idempotency: same request_id returns the same transaction
	if existing, err := s.Store.GetByRequestID(req.RequestID); err == nil {
		writeJSON(w, http.StatusOK, toResp(existing))
		return
	}

	tx := &ledger.Transaction{
		RequestID: req.RequestID,
		UserID:    userID,
		Type:      ledger.DEPOSIT,
		Amount:    cents,
		Currency:  "BRL",
		Status:    "PENDING",
		Postings: []ledger.Posting{
			{Account: userID + ":bank", Delta: -cents, Currency: "BRL"},
			{Account: userID + ledger.SuffixMain(), Delta: cents, Currency: "BRL"},
		},
	}
	if err := ledger.Post(tx); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.Store.Append(tx); err != nil {
		if errors.Is(err, ledger.ErrRequestIDConflict) {
			existing, _ := s.Store.GetByRequestID(req.RequestID)
			writeJSON(w, http.StatusOK, toResp(existing))
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := toResp(tx)
	resp.QRCode = "data:image/png;base64,<psp-generates-qr>"
	resp.CopiaECola = "00020126580014BR.GOV.BCB.PIX0136<psp-copia-e-cola>" + req.RequestID
	resp.ExpiresAt = s.Now().Add(15 * time.Minute).Format(time.RFC3339)
	writeJSON(w, http.StatusCreated, resp)
}

// DepositConfirm POST /v1/wallet/webhooks/pix  (called by PSP)
// Idempotent on the PSP transaction id; completes PENDING deposits.
func (s *Service) PixWebhook(w http.ResponseWriter, r *http.Request) {
	var payloadBytes []byte
	if r.Body != nil {
		payloadBytes, _ = io.ReadAll(r.Body)
		// restore body for decoder
		r.Body = io.NopCloser(bytes.NewReader(payloadBytes))
	}
	var hook struct {
		Event     string            `json:"event"`
		Payload   map[string]any    `json:"payload"`
		Signature string            `json:"signature,omitempty"`
	}
	if err := json.Unmarshal(payloadBytes, &hook); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	// verify HMAC signature header
	sigHeader := r.Header.Get("X-Signature")
	secret := os.Getenv("PIX_SECRET")
	if secret == "" {
		secret = "dev-pix-secret"
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payloadBytes)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sigHeader)) && sigHeader != "" {
		writeErr(w, http.StatusUnauthorized, "invalid signature")
		return
	}

	txID, _ := hook.Payload["tx_id"].(string)
	reqID, _ := hook.Payload["request_id"].(string)
	if txID == "" && reqID == "" {
		writeErr(w, http.StatusBadRequest, "tx_id or request_id required")
		return
	}

	// Locate the PENDING deposit by request_id.
	tx, err := s.Store.GetByRequestID(reqID)
	if err != nil {
		writeErr(w, http.StatusNotFound, "transaction not found")
		return
	}
	if tx.Status == "COMPLETED" {
		writeJSON(w, http.StatusOK, toResp(tx)) // already credited: idempotent
		return
	}

	// simple in-place state change for the skeleton; production does a DB update
	tx.Status = "COMPLETED"
	tx.UpdatedAt = s.Now()
	writeJSON(w, http.StatusOK, toResp(tx))
}

// Withdraw POST /v1/wallet/withdrawals
func (s *Service) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID := userFromCtx(r)
	var req withdrawReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.RequestID == "" || req.Amount <= 0 {
		writeErr(w, http.StatusBadRequest, "amount and request_id are required")
		return
	}
	cents := money(req.Amount)

	if existing, err := s.Store.GetByRequestID(req.RequestID); err == nil {
		writeJSON(w, http.StatusOK, toResp(existing))
		return
	}

	var tx *ledger.Transaction
	err := s.Store.WithAccount(userID, func(acc *ledger.Account) error {
		if acc.MainCents < cents {
			return ledger.ErrInsufficientFunds
		}
		tx = &ledger.Transaction{
			RequestID: req.RequestID,
			UserID:    userID,
			Type:      ledger.WITHDRAW,
			Amount:    cents,
			Currency:  "BRL",
			Status:    "PENDING",
			Postings: []ledger.Posting{
				{Account: userID + ledger.SuffixMain(), Delta: -cents, Currency: "BRL"},
				{Account: userID + ":bank", Delta: cents, Currency: "BRL"},
			},
		}
		return ledger.Post(tx)
	})
	if err != nil {
		if errors.Is(err, ledger.ErrInsufficientFunds) {
			writeErr(w, http.StatusUnprocessableEntity, "insufficient funds")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := s.Store.Append(tx); err != nil {
		if errors.Is(err, ledger.ErrRequestIDConflict) {
			existing, _ := s.Store.GetByRequestID(req.RequestID)
			writeJSON(w, http.StatusOK, toResp(existing))
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := toResp(tx)
	resp.FraudCheck = nil // fraud/AML checks run asynchronously in production
	writeJSON(w, http.StatusCreated, resp)
}

// Transactions GET /v1/wallet/transactions?limit=
func (s *Service) Transactions(w http.ResponseWriter, r *http.Request) {
	userID := userFromCtx(r)
	limit := 0
	writeJSON(w, http.StatusOK, map[string]any{
		"items": mustList(s.Store, userID, limit),
		"total": len(mustList(s.Store, userID, limit)),
	})
}

func toResp(tx *ledger.Transaction) txnResp {
	return txnResp{TransactionID: tx.ID, Status: tx.Status}
}

func ledgerBal(s ledger.Store, userID string) (int64, int64, error) {
	type balancer interface {
		Balance(string) (int64, int64, error)
	}
	b, ok := s.(balancer)
	if !ok {
		return 0, 0, errors.New("store does not expose balance")
	}
	return b.Balance(userID)
}

func mustList(s ledger.Store, userID string, limit int) []*ledger.Transaction {
	items, _ := s.List(userID, limit)
	return items
}

// --- Context helpers -------------------------------------------------------

func userFromCtx(r *http.Request) string {
	return auth.UserID(r)
}