-- Partial unique index: enforce email uniqueness only on active users.
-- Soft-deleted users (deleted_at IS NOT NULL) are excluded so users can
-- re-register using the same email after account deletion.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_active
ON users (email)
WHERE deleted_at IS NULL;