package entity

import "github.com/google/uuid"

type PublicProfile struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
}

type PaymentInfo struct {
	AccountNumber string `json:"account_number"`
	BankName      string `json:"bank_name"`
}

type PublicDebt struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   string    `json:"created_at"`
	ShareAmount string    `json:"share_amount"`
	Status      *string   `json:"status,omitempty"`
}

type Diagnostics struct {
	Status    string            `json:"status"`
	DBStatus  string            `json:"db_status"`
	DBStats   map[string]any    `json:"db_stats,omitempty"`
	Version   string            `json:"version"`
	UptimeSec float64           `json:"uptime_seconds"`
}
