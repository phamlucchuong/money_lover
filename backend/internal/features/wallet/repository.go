package wallet

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, wallet *Wallet) error
	GetByID(ctx context.Context, walletID uuid.UUID) (*Wallet, error)
	GetByNameAndUserID(ctx context.Context, name string, userID uuid.UUID) (*Wallet, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*Wallet, int64, error)
	Update(ctx context.Context, wallet *Wallet) error
	Delete(ctx context.Context, walletID uuid.UUID) error
}
