package wallet

import "github.com/google/uuid"

type CreateWalletRequest struct {
	Name         string     `json:"name" validate:"required,min=3,max=255"`
	WalletType   WalletType `json:"type" validate:"required,oneof=CASH BANK E_WALLET PAYLATER CREDIT_CARD"`
	Currency     Currency   `json:"currency" validate:"required, oneof=VND USD EUR JPY GBP"`
	Amount       int64      `json:"amount"`
	Notification bool       `json:"notification" validate:"required"`
	IsSummary    bool       `json:"is_summary" validate:"required"`
	IsDefault    bool       `json:"is_default" validate:"required"`
}

type UpdateWalletRequest struct {
	Name         string     `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
	WalletType   WalletType `json:"type" validate:"omitempty,oneof=CASH BANK E_WALLET PAYLATER CREDIT_CARD"`
	Currency     Currency   `json:"currency" validate:"omitempty, oneof=VND USD EUR JPY GBP"`
	Amount       int64      `json:"amount"`
	Notification bool       `json:"notification"`
	IsSummary    bool       `json:"is_summary"`
	IsDefault    bool       `json:"is_default"`
}

type WalletResponse struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	WalletType   WalletType `json:"type"`
	Currency     Currency   `json:"currency"`
	Amount       int64      `json:"amount"`
	Notification bool       `json:"notification"`
	IsDefault    bool       `json:"is_default"`
	IsSummary    bool       `json:"is_summary"`
	UserID       uuid.UUID  `json:"user_id"`
}
