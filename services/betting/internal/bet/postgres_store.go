package bet

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, dsn string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}

func (p *PostgresStore) Create(b *Bet) error {
	if b.ID != 0 {
		return nil
	}
	err := p.pool.QueryRow(context.Background(),
		`INSERT INTO bets(user_id, event_id, market, selection, odds, stake_cents, status, placed_at)
		 VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		b.UserID, b.EventID, b.Market, b.Selection, b.Odds, b.Stake, b.Status, time.Now()).
		Scan(&b.ID)
	return err
}

func (p *PostgresStore) Get(id int64) (*Bet, error) {
	var b Bet
	var settled *time.Time
	err := p.pool.QueryRow(context.Background(),
		`SELECT id, user_id, event_id, market, selection, odds, stake_cents, status, placed_at, settled_at, payout
		 FROM bets WHERE id=$1`, id).
		Scan(&b.ID, &b.UserID, &b.EventID, &b.Market, &b.Selection, &b.Odds, &b.Stake, &b.Status, &b.PlacedAt, &settled, &b.Payout)
	if err != nil {
		return nil, err
	}
	b.SettledAt = settled
	return &b, nil
}

func (p *PostgresStore) Update(b *Bet) error {
	_, err := p.pool.Exec(context.Background(),
		`UPDATE bets SET status=$1, settled_at=$2, payout=$3 WHERE id=$4`,
		b.Status, b.SettledAt, b.Payout, b.ID)
	return err
}

func (p *PostgresStore) List(userID string, status Status, limit int) ([]*Bet, error) {
	rows, err := p.pool.Query(context.Background(),
		`SELECT id, user_id, event_id, market, selection, odds, stake_cents, status, placed_at, settled_at, payout
		 FROM bets WHERE user_id=$1 AND status=$2 ORDER BY id DESC LIMIT $3`,
		userID, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Bet
	for rows.Next() {
		var b Bet
		var settled *time.Time
		if err := rows.Scan(&b.ID, &b.UserID, &b.EventID, &b.Market, &b.Selection, &b.Odds, &b.Stake, &b.Status, &b.PlacedAt, &settled, &b.Payout); err != nil {
			return nil, err
		}
		b.SettledAt = settled
		out = append(out, &b)
	}
	return out, nil
}
