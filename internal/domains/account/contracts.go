package account

// HTTP contract models used by handlers and API docs.

type CreateAccountRequest struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	AccountType   string `json:"account_type"`
	Currency      string `json:"currency"`
	AccountNumber string `json:"account_number"`
	ParentAccount string `json:"parent_account_id"`
	Color         string `json:"color"`
}
