-- 001_init.up.sql
-- Users, wallets and ledger

CREATE TABLE IF NOT EXISTS users (
    id              uuid PRIMARY KEY,
    cpf_hash        text NOT NULL UNIQUE,
    name            text NOT NULL,
    email           text UNIQUE,
    phone           text,
    birth_date      date NOT NULL,
    status          text NOT NULL DEFAULT 'PENDING', -- PENDING|ACTIVE|BLOCKED
    kyc_level       int NOT NULL DEFAULT 0,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS wallet_accounts (
    user_id         uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    main_balance_cents bigint NOT NULL DEFAULT 0,
    bonus_balance_cents bigint NOT NULL DEFAULT 0,
    version         int NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS ledger (
    id              bigserial PRIMARY KEY,
    request_id      uuid UNIQUE NOT NULL,
    user_id         uuid NOT NULL REFERENCES users(id),
    tx_type         text NOT NULL,
    amount_cents    bigint NOT NULL,
    currency        text NOT NULL DEFAULT 'BRL',
    status          text NOT NULL,
    postings        jsonb NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- 002_bets.sql
CREATE TABLE IF NOT EXISTS bets (
    id              bigserial PRIMARY KEY,
    user_id         uuid NOT NULL REFERENCES users(id),
    event_id        text NOT NULL,
    market          text NOT NULL,
    selection       text NOT NULL,
    odds            numeric(6,3) NOT NULL,
    stake_cents     bigint NOT NULL,
    status          text NOT NULL,
    placed_at       timestamptz NOT NULL DEFAULT now(),
    settled_at      timestamptz
);

-- 003_rg.sql
CREATE TABLE IF NOT EXISTS rg_limits (
    user_id         uuid NOT NULL REFERENCES users(id),
    limit_type      text NOT NULL,
    period          text NOT NULL,
    amount_cents    bigint NOT NULL,
    effective_from  timestamptz NOT NULL,
    requested_at    timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, limit_type, period)
);

CREATE TABLE IF NOT EXISTS rg_exclusions (
    user_id         uuid NOT NULL REFERENCES users(id),
    started_at      timestamptz NOT NULL DEFAULT now(),
    duration        text NOT NULL,
    until           timestamptz,
    PRIMARY KEY (user_id, started_at)
);

CREATE TABLE IF NOT EXISTS rg_reality_checks (
    id              bigserial PRIMARY KEY,
    user_id         uuid NOT NULL REFERENCES users(id),
    check_at        timestamptz NOT NULL DEFAULT now(),
    session_seconds bigint NOT NULL,
    acknowledged    bool NOT NULL
);
