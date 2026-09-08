package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/bets/bepping-platform/lib/auth"
	"github.com/bets/bepping-platform/services/betting/internal/bet"
)

// LiveMarker returns the current odds and suspension state for an event.
type LiveMarker func(eventID string) (odds float64, suspended bool)

type Service struct {
	Store bet.Store
	Wallet bet.Wallet
	RG     bet.RG
	Live   LiveMarker
	Now    func() time.Time
}

func New(s bet.Store, w bet.Wallet, rg bet.RG, live LiveMarker) *Service {
	return &Service{Store: s, Wallet: w, RG: rg, Live: live, Now: time.Now}
}

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

type placeReq struct {
	EventID   string  `json:"event_id"`
	Market    string  `json:"market"`
	Selection string  `json:"selection"`
	Odds      float64 `json:"odds"`
	Stake     float64 `json:"stake"`
	RequestID string  `json:"request_id"`
}

// Place POST /v1/bets
func (s *Service) Place(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	var req placeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.RequestID == "" {
		writeErr(w, http.StatusBadRequest, "request_id required (idempotency)")
		return
	}
	b := &bet.Bet{
		UserID:    userID,
		EventID:   req.EventID,
		Market:    req.Market,
		Selection: req.Selection,
		Odds:      req.Odds,
		Stake:     int64(req.Stake*100 + 0.5),
	}
	if err := bet.Place(s.Store, s.Wallet, s.RG, s.Live, b, req.RequestID); err != nil {
		switch {
		case errors.Is(err, bet.ErrOddsChanged):
			writeErr(w, http.StatusConflict, "odds changed")
		case errors.Is(err, bet.ErrEventSuspended):
			writeErr(w, http.StatusConflict, "event suspended")
		case errors.Is(err, bet.ErrStakeTooLarge):
			writeErr(w, http.StatusUnprocessableEntity, "stake exceeds limit")
		default:
			writeErr(w, http.StatusUnprocessableEntity, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

type settleReq struct {
	Winner bool `json:"winner"`
}

// Settle POST /v1/bets/{id}/settle  (admin/internal)
func (s *Service) Settle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid bet id")
		return
	}
	var req settleReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := bet.Settle(s.Store, s.Wallet, id, req.Winner, "settle-"+strconv.FormatInt(id, 10)+"-"+r.Header.Get("X-Replay")); err != nil {
		writeErr(w, http.StatusConflict, err.Error())
		return
	}
	b, _ := s.Store.Get(id)
	writeJSON(w, http.StatusOK, b)
}

type cashoutReq struct {
	Offer float64 `json:"offer"`
}

// Cashout POST /v1/bets/{id}/cashout
func (s *Service) Cashout(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid bet id")
		return
	}
	var req cashoutReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := bet.Cashout(s.Store, s.Wallet, id, int64(req.Offer*100+0.5), "cashout-"+strconv.FormatInt(id, 10)); err != nil {
		writeErr(w, http.StatusConflict, err.Error())
		return
	}
	b, _ := s.Store.Get(id)
	writeJSON(w, http.StatusOK, b)
}

// List GET /v1/bets?status=
func (s *Service) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserID(r)
	status := bet.Status(r.URL.Query().Get("status"))
	items, err := s.Store.List(userID, status, 50)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}