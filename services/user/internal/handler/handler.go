package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"time"

	"github.com/bets/bepping-platform/lib/auth"
	"github.com/bets/bepping-platform/services/user/internal/rg"
)

type Service struct {
	RG rg.Store
	Now func() time.Time
}

func New(s rg.Store) *Service {
	return &Service{RG: s, Now: time.Now}
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

// --- context (mirror of wallet-service; replace with shared auth lib) ----

func UserID(r *http.Request) string {
	return auth.UserID(r)
}

// --- Responsible Gaming handlers ------------------------------------------

// GetLimits GET /v1/rg/limits
func (s *Service) GetLimits(w http.ResponseWriter, r *http.Request) {
	user := UserID(r)
	limits, err := s.RG.Limits(user)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if limits == nil {
		limits = []rg.Limit{}
	}
	writeJSON(w, http.StatusOK, limits)
}

type setLimitReq struct {
	LimitType       string  `json:"limit_type"`
	Period          string  `json:"period"`
	Amount          float64 `json:"amount"`
	DurationMinutes int64   `json:"duration_minutes"`
}

// SetLimit PUT /v1/rg/limits
// Down => immediate. Up => effective_from = now + 24h.
func (s *Service) SetLimit(w http.ResponseWriter, r *http.Request) {
	user := UserID(r)
	var req setLimitReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	lt := rg.LimitType(req.LimitType)
	switch lt {
	case rg.DEPOSIT, rg.LOSS, rg.SESSION_TIME:
	default:
		writeErr(w, http.StatusBadRequest, "invalid limit_type")
		return
	}
	p := rg.Period(req.Period)
	switch p {
	case rg.DAILY, rg.WEEKLY, rg.MONTHLY:
	default:
		writeErr(w, http.StatusBadRequest, "invalid period")
		return
	}

	now := s.Now().UTC()

	var amount int64
	if lt == rg.SESSION_TIME {
		amount = req.DurationMinutes
		if amount <= 0 {
			writeErr(w, http.StatusBadRequest, "duration_minutes required for SESSION_TIME")
			return
		}
	} else {
		amount = int64(math.Round(req.Amount * 100))
		if amount <= 0 {
			writeErr(w, http.StatusBadRequest, "amount must be positive")
			return
		}
	}

	// Compare with current configured limit to decide effective_from.
	effective := now
	existing, _ := s.RG.Limits(user)
	for _, l := range existing {
		if l.Type == lt && l.Period == p {
			if amount > l.Amount {
				effective = now.Add(24 * time.Hour)
			}
			break
		}
	}

	l := rg.Limit{
		Type:          lt,
		Period:        p,
		Amount:        amount,
		EffectiveFrom: effective,
		RequestedAt:   now,
	}
	if err := s.RG.SetLimit(user, l); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"effective_from": effective.Format(time.RFC3339),
	})
}

// SelfExclude POST /v1/rg/self-exclusion
func (s *Service) SelfExclude(w http.ResponseWriter, r *http.Request) {
	user := UserID(r)
	var req struct {
		Duration string `json:"duration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	dur, permanent, err := rg.ParseDuration(req.Duration)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	now := s.Now().UTC()
	ex := rg.Exclusion{
		UserID:    user,
		Start:     now,
		Duration:  req.Duration,
		Permanent: permanent,
	}
	if !permanent {
		ex.Until = now.Add(dur)
	}
	if err := s.RG.BeginExclusion(user, ex); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	until := "PERMANENT"
	if !permanent {
		until = ex.Until.Format(time.RFC3339)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "EXCLUDED",
		"until":  until,
	})
}

type realityCheckReq struct {
	SessionSeconds int64 `json:"session_seconds"`
	NetResult      int64 `json:"net_result"`
}

// RealityCheck POST /v1/rg/reality-check
func (s *Service) RealityCheck(w http.ResponseWriter, r *http.Request) {
	user := UserID(r)
	var req realityCheckReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := s.RG.RegisterRealityCheck(user, rg.RealityCheck{
		UserID:         user,
		At:             s.Now().UTC(),
		SessionSeconds: req.SessionSeconds,
		Ack:            true,
	}); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"acknowledged":   true,
		"session_seconds": req.SessionSeconds,
		"net_result":     float64(req.NetResult) / 100,
	})
}