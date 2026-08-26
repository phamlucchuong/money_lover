package wallet

import "github.com/google/uuid"

type CreateWalletRequest struct {
	Name         string     `json:"name" binding:"required"`
	WalletType   WalletType `json:"type" binding:"required,oneof=CASH BANK E-WALLET PAYLATER CREDIT_CARD"`
	Currency     Currency   `json:"currency" binding:"required"`
	Amount       int64      `json:"amount" binding:"required"`
	Notification bool       `json:"notification"`
	IsSummary    bool       `json:"is_summary"`
}

type UpdateWalletRequest struct {
	Name         string     `json:"name" binding:"required"`
	WalletType   WalletType `json:"type" binding:"required,oneof=CASH BANK E-WALLET PAYLATER CREDIT_CARD"`
	Currency     Currency   `json:"currency" binding:"required"`
	Amount       int64      `json:"amount"`
	Notification bool       `json:"notification"`
	IsSummary    bool       `json:"is_summary"`
}

type WalletResponse struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	WalletType   WalletType `json:"type"`
	Currency     Currency   `json:"currency"`
	Amount       int64      `json:"amount"`
	Notification bool       `json:"notification"`
	IsSummary    bool       `json:"is_summary"`
	UserID       uuid.UUID  `json:"user_id"`
}
