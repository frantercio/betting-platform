package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bets/bepping-platform/lib/auth"
	"github.com/bets/bepping-platform/services/user/internal/rg"
)

func fixedNow() time.Time { return time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC) }

func TestSetDepositLimitDownIsImmediate(t *testing.T) {
	store := rg.NewMemStore()
	svc := New(store)
	svc.Now = fixedNow
	user := "user_rg1"

	// seed a higher existing limit
	store.SetLimit(user, rg.Limit{Type: rg.DEPOSIT, Period: rg.DAILY, Amount: 100000, EffectiveFrom: fixedNow()})

	rec := callSetLimit(svc, user, `{"limit_type":"DEPOSIT","period":"DAILY","amount":50.00}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}
	// down should be effective immediately
	var resp struct {
		EffectiveFrom string `json:"effective_from"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.EffectiveFrom != fixedNow().Format(time.RFC3339) {
		t.Fatalf("down-limit should be immediate, got %s", resp.EffectiveFrom)
	}
}

func TestSetDepositLimitUpHas24hDelay(t *testing.T) {
	store := rg.NewMemStore()
	svc := New(store)
	svc.Now = fixedNow
	user := "user_rg2"

	store.SetLimit(user, rg.Limit{Type: rg.DEPOSIT, Period: rg.DAILY, Amount: 5000, EffectiveFrom: fixedNow()})

	rec := callSetLimit(svc, user, `{"limit_type":"DEPOSIT","period":"DAILY","amount":200.00}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		EffectiveFrom string `json:"effective_from"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	want := fixedNow().Add(24 * time.Hour).Format(time.RFC3339)
	if resp.EffectiveFrom != want {
		t.Fatalf("up-limit should delay 24h: got %s want %s", resp.EffectiveFrom, want)
	}
}

func TestSelfExclusionBlocks(t *testing.T) {
	store := rg.NewMemStore()
	svc := New(store)
	svc.Now = fixedNow
	user := "user_rg3"

	rec := callSelfExclude(svc, user, "30D")
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}

	ex, active, err := rg.Active(user, store, fixedNow())
	if err != nil || !active {
		t.Fatalf("expected active exclusion: active=%v err=%v", active, err)
	}
	if ex.Duration != "30D" {
		t.Fatalf("expected 30D exclusion, got %s", ex.Duration)
	}
	// after expiry, not active
	expired := fixedNow().Add(35 * 24 * time.Hour)
	if _, active, _ := rg.Active(user, store, expired); active {
		t.Fatal("exclusion should have expired")
	}
}

func callSetLimit(svc *Service, user, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPut, "/v1/rg/limits", strings.NewReader(body))
	rec := httptest.NewRecorder()
	svc.SetLimit(rec, auth.WithUser(req, user))
	return rec
}

func callSelfExclude(svc *Service, user, dur string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/rg/self-exclusion", strings.NewReader(`{"duration":"`+dur+`"}`))
	rec := httptest.NewRecorder()
	svc.SelfExclude(rec, auth.WithUser(req, user))
	return rec
}