package wallet

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WalletType string

const (
	WalletTypeCash       WalletType = "CASH"
	WalletTypeBank       WalletType = "BANK"
	WalletTypeEWallet    WalletType = "E_WALLET"
	WalletTypePaylater   WalletType = "PAYLATER"
	WalletTypeCreditCard WalletType = "CREDIT_CARD"
)

type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyVND Currency = "VND"
	CurrencyJPY Currency = "JPY"
	CurrencyGBP Currency = "GBP"
)

type Wallet struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name         string         `gorm:"type:varchar(255);not null"`
	WalletType   WalletType     `gorm:"type:varchar(32);not null;default:'CASH'"`
	Currency     Currency       `gorm:"type:varchar(32);not null;default:'VND'"`
	Amount       int64          `gorm:"type:bigint;not null;default:0"`
	IsDefault    bool           `gorm:"type:boolean;not null;default:false"`
	Notification bool           `gorm:"type:boolean;not null;default:false"`
	IsSummary    bool           `gorm:"type:boolean;not null;default:false"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null"`
	CreatedAt    time.Time      `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt    time.Time      `gorm:"type:timestamptz;not null;default:now()"`
	DeletedAt    gorm.DeletedAt `gorm:"type:timestamptz;default:null"`
}

func (Wallet) TableName() string {
	return "wallets"
}
