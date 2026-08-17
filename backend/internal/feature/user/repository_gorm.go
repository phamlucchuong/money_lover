package user

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) WithTx(ctx context.Context, fn func(repo Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &repository{db: tx}
		return fn(txRepo)
	})
}

func (r *repository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *repository) GetByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *repository) GetByEmailAndDeletedAtIsNull(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("email = ? AND deleted_at is NULL", email).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *repository) GetAll(ctx context.Context) ([]*User, error) {
	var users []*User
	err := r.db.WithContext(ctx).Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *repository) Update(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Updates(user).Error
}

func (r *repository) Delete(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", userID).Update("deleted_at", gorm.Expr("NOW()")).Error
}
