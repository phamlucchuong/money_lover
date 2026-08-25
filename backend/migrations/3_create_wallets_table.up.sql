CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    wallet_type VARCHAR(32) NOT NULL DEFAULT 'CASH' CHECK (wallet_type IN ('CASH', 'BANK', 'CREDIT_CARD', 'E_WALLET', 'PAYLATER')),
    currency VARCHAR(32) NOT NULL DEFAULT 'VND' CHECK (currency IN ('VND', 'USD', 'EUR', 'JPY', 'GBP')),
    amount BIGINT NOT NULL DEFAULT 0,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    notification BOOLEAN NOT NULL DEFAULT FALSE,
    is_summary BOOLEAN NOT NULL DEFAULT FALSE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Speed up per-user wallet lookups; deleted_at also supports the partial
-- unique index below.
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets (user_id);
CREATE INDEX IF NOT EXISTS idx_wallets_deleted_at ON wallets (deleted_at);

-- Partial unique index: enforce (user_id, name) uniqueness only on active
-- wallets. Soft-deleted wallets (deleted_at IS NOT NULL) are excluded so
-- users can reuse a wallet name after deletion.
CREATE UNIQUE INDEX IF NOT EXISTS idx_wallets_user_id_name_active
ON wallets (user_id, name)
WHERE deleted_at IS NULL;
