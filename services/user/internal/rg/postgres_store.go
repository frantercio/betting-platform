package rg

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
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

func (p *PostgresStore) SetLimit(userID string, l Limit) error {
	_, err := p.pool.Exec(context.Background(),
		`INSERT INTO rg_limits(user_id, limit_type, period, amount_cents, effective_from, requested_at)
		 VALUES($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (user_id, limit_type, period)
		 DO UPDATE SET amount_cents=$4, effective_from=$5, requested_at=$6`,
		userID, l.Type, l.Period, l.Amount, l.EffectiveFrom, l.RequestedAt)
	return err
}

func (p *PostgresStore) Limits(userID string) ([]Limit, error) {
	rows, err := p.pool.Query(context.Background(),
		`SELECT limit_type, period, amount_cents, effective_from, requested_at
		 FROM rg_limits WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Limit
	for rows.Next() {
		var l Limit
		if err := rows.Scan(&l.Type, &l.Period, &l.Amount, &l.EffectiveFrom, &l.RequestedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}

func (p *PostgresStore) ActiveExclusion(userID string) (*Exclusion, error) {
	var e Exclusion
	err := p.pool.QueryRow(context.Background(),
		`SELECT user_id, started_at, duration, until FROM rg_exclusions
		 WHERE user_id=$1 ORDER BY started_at DESC LIMIT 1`, userID).
		Scan(&e.UserID, &e.Start, &e.Duration, &e.Until)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	e.Permanent = e.Duration == "PERMANENT"
	return &e, nil
}

func (p *PostgresStore) BeginExclusion(userID string, e Exclusion) error {
	until := e.Until
	if e.Permanent {
		until = time.Time{}
	}
	_, err := p.pool.Exec(context.Background(),
		`INSERT INTO rg_exclusions(user_id, started_at, duration, until)
		 VALUES($1,$2,$3,$4)`,
		userID, e.Start, e.Duration, until)
	return err
}

func (p *PostgresStore) RegisterRealityCheck(userID string, rc RealityCheck) error {
	_, err := p.pool.Exec(context.Background(),
		`INSERT INTO rg_reality_checks(user_id, check_at, session_seconds, acknowledged)
		 VALUES($1,$2,$3,$4)`,
		userID, rc.At, rc.SessionSeconds, rc.Ack)
	return err
}
