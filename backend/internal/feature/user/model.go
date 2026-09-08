package user

import (
	"time"

	"chuongpl/quan-ly-chi-tieu/internal/features/wallet"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string          `gorm:"type:varchar(255);not null"`
	Email     string          `gorm:"type:varchar(255);not null"`
	Password  string          `gorm:"type:varchar(72);not null"`
	Wallets   []wallet.Wallet `gorm:"foreignKey:UserID;references:ID"`
	CreatedAt time.Time       `gorm:"type:timestamptz;autoCreateTime;not null;default:now()"`
	UpdatedAt time.Time       `gorm:"type:timestamptz;autoUpdateTime;not null;default:now()"`
	DeletedAt gorm.DeletedAt  `gorm:"index;type:timestamptz;default:null"`
}

func (User) TableName() string {
	return "users"
}
