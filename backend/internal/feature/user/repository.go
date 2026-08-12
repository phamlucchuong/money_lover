package user

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RepositoryInterface interface {
	WithTx(ctx context.Context, fn func(repo RepositoryInterface) error) error
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error)
	GetUserByEmailAndDeletedAtIsNull(ctx context.Context, email string) (*User, error)
	GetAll(ctx context.Context) ([]*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithTx(ctx context.Context, fn func(repo RepositoryInterface) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &Repository{db: tx}
		return fn(txRepo)
	})
}

func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *Repository) GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetUserByEmailAndDeletedAtIsNull(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("email = ? AND deleted_at is NULL", email).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) GetAll(ctx context.Context) ([]*User, error) {
	var users []*User
	err := r.db.WithContext(ctx).Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Updates(user).Error
}

func (r *Repository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", userID).Update("deleted_at", gorm.Expr("NOW()")).Error
}
