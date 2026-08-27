package wallet

import (
	"context"
	"errors"
	"log/slog"
	"math"

	"chuongpl/quan-ly-chi-tieu/internal/config"
	"chuongpl/quan-ly-chi-tieu/internal/pkg"
	"chuongpl/quan-ly-chi-tieu/internal/platform/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrWalletAlreadyExists = errors.New("wallet already exists")
)

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
	existing, err := s.repo.GetByUserAndName(ctx, userID, req.Name)
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

func (s *service) GetByID(ctx context.Context, userID, walletID uuid.UUID) (*WalletResponse, error) {
	wallet, err := s.repo.GetByUserAndID(ctx, userID, walletID)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrWalletNotFound
		default:
			return nil, err
		}
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

func (s *service) GetAllByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*WalletResponse, *pkg.PaginationMeta, error) {
	offset := (page - 1) * pageSize
	wallets, total, err := s.repo.GetAllByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, nil, err
	}

	responses := make([]*WalletResponse, 0, len(wallets))
	for _, wallet := range wallets {
		responses = append(responses, &WalletResponse{
			ID:           wallet.ID,
			Name:         wallet.Name,
			WalletType:   wallet.WalletType,
			Currency:     wallet.Currency,
			Amount:       wallet.Amount,
			Notification: wallet.Notification,
			IsSummary:    wallet.IsSummary,
			UserID:       wallet.UserID,
		})
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	meta := &pkg.PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: int(total),
		TotalPages: totalPages,
	}

	return responses, meta, nil
}

func (s *service) Update(ctx context.Context, userID, walletID uuid.UUID, req UpdateWalletRequest) (*WalletResponse, error) {
	wallet, err := s.repo.GetByUserAndID(ctx, userID, walletID)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, ErrWalletNotFound
		default:
			return nil, err
		}
	}

	if req.Name != "" {
		wallet.Name = req.Name
	}

	if req.WalletType != "" {
		wallet.WalletType = req.WalletType
	}

	if req.Currency != "" {
		wallet.Currency = req.Currency
	}

	if req.Amount != 0 {
		wallet.Amount = req.Amount
	}

	wallet.Notification = req.Notification
	wallet.IsSummary = req.IsSummary
	wallet.IsDefault = req.IsDefault

	if err := s.repo.Update(ctx, wallet); err != nil {
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
		IsDefault:    wallet.IsDefault,
		UserID:       wallet.UserID,
	}, nil
}

func (s *service) Delete(ctx context.Context, userID, walletID uuid.UUID) error {
	err := s.repo.DeleteByUserID(ctx, userID, walletID)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return ErrWalletNotFound
		default:
			return err
		}
	}
	return nil
}
