package db

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestIsUniqueViolation(t *testing.T) {
	testCases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "unique violation pg error",
			err:  &pgconn.PgError{Code: "23505"},
			want: true,
		},
		{
			name: "wrapped unique violation",
			err:  fmt.Errorf("create user: %w", &pgconn.PgError{Code: "23505"}),
			want: true,
		},
		{
			name: "other pg error code",
			err:  &pgconn.PgError{Code: "23502"},
			want: false,
		},
		{
			name: "non-pg error",
			err:  errors.New("connection refused"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsUniqueViolation(tc.err))
		})
	}
}