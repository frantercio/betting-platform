package bet

import "sync"

var _ Store = (*MemStore)(nil)

// MemStore is an in-memory Store for dev/tests.
type MemStore struct {
	mu   sync.RWMutex
	bets map[int64]*Bet
	seq  int64
}

func NewMemStore() *MemStore {
	return &MemStore{bets: map[int64]*Bet{}}
}

func (m *MemStore) Create(b *Bet) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	b.ID = m.seq
	m.bets[b.ID] = b
	return nil
}

func (m *MemStore) Get(id int64) (*Bet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, ok := m.bets[id]
	if !ok {
		return nil, ErrNotOpen
	}
	cp := *b
	return &cp, nil
}

func (m *MemStore) Update(b *Bet) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.bets[b.ID]; !ok {
		return ErrNotOpen
	}
	cp := *b
	m.bets[b.ID] = &cp
	return nil
}

func (m *MemStore) List(userID string, status Status, limit int) ([]*Bet, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []*Bet{}
	for _, b := range m.bets {
		if b.UserID != userID || (status != "" && b.Status != status) {
			continue
		}
		cp := *b
		out = append(out, &cp)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}