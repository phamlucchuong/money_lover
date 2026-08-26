package wallet

import (
	"context"
	"errors"
	"log/slog"

	"chuongpl/quan-ly-chi-tieu/internal/config"
	"chuongpl/quan-ly-chi-tieu/internal/platform/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrWalletAlreadyExists = errors.New("wallet already exists")

type Service interface{}

type service struct {
	repo Repository
	cfg  *config.Config
	log  *slog.Logger
}

func NewService(repo Repository, cfg *config.Config, log *slog.Logger) Service {
	return &service{
		repo: repo,
		cfg:  cfg,
		log:  log,
	}
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, req CreateWalletRequest) (*WalletResponse, error) {
	existing, err := s.repo.GetByNameAndUserID(ctx, req.Name, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if existing != nil {
		return nil, ErrWalletAlreadyExists
	}

	wallet := &Wallet{
		Name:         req.Name,
		WalletType:   req.WalletType,
		Currency:     req.Currency,
		Amount:       req.Amount,
		Notification: req.Notification,
		IsSummary:    req.IsSummary,
		UserID:       userID,
	}

	if err = s.repo.Create(ctx, wallet); err != nil {
		if db.IsUniqueViolation(err) {
			return nil, ErrWalletAlreadyExists
		}
		return nil, err
	}

	return &WalletResponse{
		ID:           wallet.ID,
		Name:         wallet.Name,
		WalletType:   wallet.WalletType,
		Currency:     wallet.Currency,
		Amount:       wallet.Amount,
		Notification: wallet.Notification,
		IsSummary:    wallet.IsSummary,
		UserID:       wallet.UserID,
	}, nil
}

func (s *service) GetAllByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*WalletResponse, error) {
	var wallets []*Wallet
	wallets, _, err := s.repo.GetAllByUserID(ctx, userID, offset, limit)
	if err != nil {
		return nil, err
	}

	responses := make([]*WalletResponse, len(wallets))
	for i, wallet := range wallets {
		responses[i] = &WalletResponse{
			ID:           wallet.ID,
			Name:         wallet.Name,
			WalletType:   wallet.WalletType,
			Currency:     wallet.Currency,
			Amount:       wallet.Amount,
			Notification: wallet.Notification,
			IsSummary:    wallet.IsSummary,
			UserID:       wallet.UserID,
		}
	}

	return responses, nil
}
