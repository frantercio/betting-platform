package rg

import (
	"fmt"
	"sync"
	"time"
)

// LimitType and Period follow the OpenAPI RG schemas.
type LimitType string
type Period string

const (
	DEPOSIT     LimitType = "DEPOSIT"
	LOSS        LimitType = "LOSS"
	SESSION_TIME LimitType = "SESSION_TIME"
)

const (
	DAILY   Period = "DAILY"
	WEEKLY  Period = "WEEKLY"
	MONTHLY Period = "MONTHLY"
)

// Limit is one configured user limit.
// Changing DOWN takes effect immediately; changing UP honours a 24h delay
// (effective_from) to prevent impulse decisions.
type Limit struct {
	Type        LimitType `json:"limit_type"`
	Period      Period    `json:"period"`
	Amount      int64     `json:"amount"` // cents; minutes for SESSION_TIME
	EffectiveFrom time.Time `json:"effective_from"`
	RequestedAt time.Time `json:"requested_at"`
}

// Store persists limits and exclusions (Postgres in prod).
type Store interface {
	SetLimit(userID string, l Limit) error
	Limits(userID string) ([]Limit, error)
	ActiveExclusion(userID string) (*Exclusion, error)
	BeginExclusion(userID string, e Exclusion) error
	RegisterRealityCheck(userID string, rc RealityCheck) error
}

// Exclusion models temporary or permanent self-exclusion.
type Exclusion struct {
	UserID    string    `json:"user_id"`
	Start     time.Time `json:"started_at"`
	Duration  string    `json:"duration"` // 24H|7D|30D|90D|6M|PERMANENT
	Until     time.Time `json:"until"`
	Permanent bool      `json:"permanent"`
}

// RealityCheck is logged whenever the 60-min popup is shown/acknowledged.
type RealityCheck struct {
	UserID  string    `json:"user_id"`
	At      time.Time `json:"check_at"`
	SessionSeconds int64 `json:"session_seconds"`
	Ack     bool      `json:"acknowledged"`
}

// Active reports whether user (not) currently excluded.
func Active(userID string, s Store, now time.Time) (*Exclusion, bool, error) {
	ex, err := s.ActiveExclusion(userID)
	if err != nil {
		return nil, false, err
	}
	if ex == nil {
		return nil, false, nil
	}
	if !ex.Permanent && now.After(ex.Until) {
		return nil, false, nil // expired
	}
	return ex, true, nil
}

// WithinDepositLimit sums completed deposits for the period and checks cap.
// In production, sum from the ledger events (Kafka) rather than a counter.
func WithinDepositLimit(limit int64, spentInPeriod int64) bool {
	return spentInPeriod < limit
}

var _ Store = (*MemStore)(nil)

// MemStore is an in-memory Store for dev/tests.
type MemStore struct {
	mu       sync.Mutex
	limits   map[string][]Limit
	excl     map[string][]Exclusion
	rechecks []RealityCheck
}

func NewMemStore() *MemStore {
	return &MemStore{limits: map[string][]Limit{}, excl: map[string][]Exclusion{}}
}

func (m *MemStore) SetLimit(userID string, l Limit) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ls := m.limits[userID]
	for i := range ls {
		if ls[i].Type == l.Type && ls[i].Period == l.Period {
			ls[i] = l
			m.limits[userID] = ls
			return nil
		}
	}
	m.limits[userID] = append(ls, l)
	return nil
}

func (m *MemStore) Limits(userID string) ([]Limit, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Limit(nil), m.limits[userID]...), nil
}

func (m *MemStore) ActiveExclusion(userID string) (*Exclusion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	es := m.excl[userID]
	if len(es) == 0 {
		return nil, nil
	}
	last := es[len(es)-1]
	return &last, nil
}

func (m *MemStore) BeginExclusion(userID string, e Exclusion) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.excl[userID] = append(m.excl[userID], e)
	return nil
}

func (m *MemStore) RegisterRealityCheck(userID string, rc RealityCheck) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rechecks = append(m.rechecks, rc)
	return nil
}

// ParseDuration maps 24H..PERMANENT to time.Duration.
func ParseDuration(s string) (time.Duration, bool, error) {
	switch s {
	case "24H":
		return 24 * time.Hour, false, nil
	case "7D":
		return 7 * 24 * time.Hour, false, nil
	case "30D":
		return 30 * 24 * time.Hour, false, nil
	case "90D":
		return 90 * 24 * time.Hour, false, nil
	case "6M":
		return 182 * 24 * time.Hour, false, nil
	case "PERMANENT":
		return 0, true, nil
	default:
		return 0, false, fmt.Errorf("unknown duration: %s", s)
	}
}