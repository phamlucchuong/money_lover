package wallet

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(ctx context.Context, wallet *Wallet) error {
	return r.db.WithContext(ctx).Create(wallet).Error
}

func (r *repository) GetByUserAndID(ctx context.Context, userID, walletID uuid.UUID) (*Wallet, error) {
	var wallet Wallet
	err := r.db.WithContext(ctx).Model(&Wallet{}).Where("id = ? AND user_id = ? AND deleted_at IS NULL", walletID, userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

func (r *repository) GetByUserAndName(ctx context.Context, userID uuid.UUID, name string) (*Wallet, error) {
	var wallet Wallet
	err := r.db.WithContext(ctx).Model(&Wallet{}).Where("name = ? AND user_id = ? AND deleted_at IS NULL", name, userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

func (r *repository) GetAllByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*Wallet, int64, error) {
	var wallets []*Wallet
	var total int64

	err := r.db.WithContext(ctx).Model(&Wallet{}).Where("user_id = ? AND deleted_at IS NULL", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).Model(&Wallet{}).Where("user_id = ? AND deleted_at IS NULL", userID).Offset(offset).Limit(limit).Find(&wallets).Order("created_at DESC").Error
	if err != nil {
		return nil, 0, err
	}

	return wallets, total, nil
}

func (r *repository) Update(ctx context.Context, wallet *Wallet) error {
	result := r.db.WithContext(ctx).Model(&Wallet{}).Where("id = ? AND deleted_at IS NULL", wallet.ID).Updates(wallet)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *repository) DeleteByUserID(ctx context.Context, userID, walletID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Wallet{}, "id = ? AND user_id = ? AND deleted_at IS NULL", walletID, userID)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
