package db

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Postgres error codes
const (
	pgUniqueViolation = "23505"
)

// IsUniqueViolation checks whether an error is a Postgres unique
// constraint violation (SQLSTATE 23505). Returns true when the error
// (or any error in its chain) is a *pgconn.PgError with that code.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}