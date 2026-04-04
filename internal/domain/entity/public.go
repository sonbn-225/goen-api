package entity

type PublicProfile struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}

type PaymentInfo struct {
	AccountNumber string `json:"account_number"`
	BankName      string `json:"bank_name"`
}

type Diagnostics struct {
	Status    string            `json:"status"`
	DBStatus  string            `json:"db_status"`
	DBStats   map[string]any    `json:"db_stats,omitempty"`
	Version   string            `json:"version"`
	UptimeSec float64           `json:"uptime_seconds"`
}
