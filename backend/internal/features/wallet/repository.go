package wallet

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, wallet *Wallet) error
	GetByUserAndID(ctx context.Context, userID, walletID uuid.UUID) (*Wallet, error)
	GetByUserAndName(ctx context.Context, userID uuid.UUID, name string) (*Wallet, error)
	GetAllByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*Wallet, int64, error)
	Update(ctx context.Context, wallet *Wallet) error
	DeleteByUserID(ctx context.Context, userID, walletID uuid.UUID) error
}
