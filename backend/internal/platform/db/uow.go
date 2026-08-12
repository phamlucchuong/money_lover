package db

import (
	"context"

	"gorm.io/gorm"
)

type UnitOfWork interface {
	Do(ctx context.Context, fn func(tx *gorm.DB) error) error
}

type gormUnitOfWork struct {
	db *gorm.DB
}

func NewGormUnitOfWork(db *gorm.DB) UnitOfWork {
	return &gormUnitOfWork{db: db}
}

func (uow *gormUnitOfWork) Do(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return uow.db.WithContext(ctx).Transaction(fn)

}
