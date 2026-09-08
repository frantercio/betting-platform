package ledger

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrRequestIDConflict  = errors.New("request_id already used")
	ErrAccountNotFound    = errors.New("account not found")
)

// TxnType categorizes every movement. Double-entry postings split a single
// logical transaction into debits and credits between accounts.
type TxnType string

const (
	DEPOSIT   TxnType = "DEPOSIT"
	WITHDRAW  TxnType = "WITHDRAW"
	BET_DEBIT TxnType = "BET_DEBIT"
	BET_CREDIT TxnType = "BET_CREDIT"
	BONUS     TxnType = "BONUS"
	ADJUST    TxnType = "ADJUST"
	REVERSAL  TxnType = "REVERSAL"
)

// Posting is one side of a double-entry entry.
type Posting struct {
	Account  string  `json:"account"`
	Delta    int64   `json:"delta"` // in cents, "cêntimos", integer math only
	Currency string  `json:"currency"`
}

// Transaction is the public event. status transitions PENDING->COMPLETED/FAILED.
type Transaction struct {
	ID        int64       `json:"id"`
	RequestID string      `json:"request_id"` // idempotency key
	UserID    string      `json:"user_id"`
	Type      TxnType     `json:"type"`
	Amount    int64       `json:"amount"` // positive cents
	Currency  string      `json:"currency"`
	Status    string      `json:"status"` // PENDING | COMPLETED | FAILED
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Postings  []Posting   `json:"postings"`
}

// Store is the repository interface (swap for Postgres in production).
type Store interface {
	WithAccount(userID string, fn func(*Account) error) error
	Append(tx *Transaction) error
	GetByRequestID(requestID string) (*Transaction, error)
	List(userID string, limit int) ([]*Transaction, error)
}

// Account holds the running balance for one user.
type Account struct {
	UserID string
	MainCents int64 // real money, withdrawn cash
	BonusCents int64
	version int
}

var _ Store = (*MemStore)(nil)

// MemStore is an in-memory, mutex-guarded Store for development and tests.
// Replace with a relational ledger (append-only) in production.
type MemStore struct {
	mu      sync.RWMutex
	accs    map[string]*Account
	txs     []*Transaction
	byReq   map[string]*Transaction
	seq     int64
	now     func() time.Time
}

func NewMemStore() *MemStore {
	return &MemStore{
		accs:  map[string]*Account{},
		txs:   []*Transaction{},
		byReq: map[string]*Transaction{},
		now:   time.Now,
	}
}

func mainAccount(userID string) string { return userID + ":main" }

func (m *MemStore) WithAccount(userID string, fn func(*Account) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	acc := m.accountLocked(mainAccount(userID))
	return fn(acc)
}

// accountLocked returns the account for a named account id, creating it if
// missing. Caller must hold the lock.
func (m *MemStore) accountLocked(name string) *Account {
	acc, ok := m.accs[name]
	if !ok {
		acc = &Account{UserID: name}
		m.accs[name] = acc
	}
	return acc
}

func (m *MemStore) Append(tx *Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, dup := m.byReq[tx.RequestID]; dup {
		return ErrRequestIDConflict
	}
	m.seq++
	tx.ID = m.seq
	now := m.now()
	tx.CreatedAt = now
	tx.UpdatedAt = now
	for _, p := range tx.Postings {
		acc := m.accountLocked(p.Account)
		acc.MainCents += p.Delta
		acc.version++
	}
	m.txs = append(m.txs, tx)
	m.byReq[tx.RequestID] = tx
	return nil
}

// SuffixMain keeps the account naming convention (userID:main in prod).
func SuffixMain() string { return ":main" }

func (m *MemStore) GetByRequestID(requestID string) (*Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	tx, ok := m.byReq[requestID]
	if !ok {
		return nil, ErrRequestIDConflict
	}
	return tx, nil
}

func (m *MemStore) List(userID string, limit int) ([]*Transaction, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Transaction, 0, limit)
	for i := len(m.txs) - 1; i >= 0 && len(out) < limit; i-- {
		if m.txs[i].UserID == userID {
			out = append(out, m.txs[i])
		}
	}
	return out, nil
}

func (m *MemStore) Balance(userID string) (main, bonus int64, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	acc, ok := m.accs[mainAccount(userID)]
	if !ok {
		return 0, 0, ErrAccountNotFound
	}
	return acc.MainCents, acc.BonusCents, nil
}

// Post creates a validated double-entry posting pair for a transaction.
func Post(tx *Transaction) error {
	if tx.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if tx.Type == "" {
		return fmt.Errorf("type is required")
	}
	if tx.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if len(tx.Postings) != 2 {
		return fmt.Errorf("internal: double-entry requires exactly 2 postings")
	}
	var net int64
	for _, p := range tx.Postings {
		net += p.Delta
	}
	if net != 0 {
		return fmt.Errorf("postings must balance to zero: net=%d", net)
	}
	return nil
}