package user

import (
	"context"

	"github.com/google/uuid"
)

//go:generate mockery
type Repository interface {
	WithTx(ctx context.Context, fn func(repo Repository) error) error
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
	GetByEmailAndDeletedAtIsNull(ctx context.Context, email string) (*User, error)
	GetAll(ctx context.Context, offset, limit int) ([]*User, int64, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, userID uuid.UUID) error
}
