package dto

type CreateAccountRequest struct {
	Name            string  `json:"name"`
	AccountNumber   *string `json:"account_number,omitempty"`
	Color           *string `json:"color,omitempty"`
	AccountType     string  `json:"account_type"`
	Currency        string  `json:"currency"`
	ParentAccountID *string `json:"parent_account_id,omitempty"`
}

type UpsertShareRequest struct {
	Login      string `json:"login"`
	Permission string `json:"permission"`
}
