package postgres

const (
	// AccountColumnsSQL represents the standard column selection for the accounts table.
	AccountColumnsSQL = `a.id, a.name, a.account_number, a.account_type, a.currency, a.parent_account_id, a.status, a.settings, a.closed_at, a.created_at, a.updated_at, a.deleted_at`

	// AccountBalanceSQL represents the complex balance calculation logic used across multiple repo methods.
	AccountBalanceSQL = `COALESCE(SUM(
		CASE
			WHEN t.type = 'income' AND t.account_id = a.id THEN t.amount
			WHEN t.type = 'expense' AND t.account_id = a.id THEN -t.amount
			WHEN t.type = 'transfer' AND t.to_account_id = a.id THEN COALESCE(t.to_amount, t.amount)
			WHEN t.type = 'transfer' AND t.from_account_id = a.id THEN -COALESCE(t.from_amount, t.amount)
			ELSE 0
		END
	), 0)::text AS balance`

	// TransactionColumnsSQL represents the standard column selection for the transactions table.
	TransactionColumnsSQL = `t.id, t.external_ref, t.type, t.occurred_at, to_char(t.occurred_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS occurred_date, t.amount::text, t.from_amount::text, t.to_amount::text, (SELECT li.note FROM transaction_line_items li WHERE li.transaction_id = t.id ORDER BY li.id LIMIT 1) AS description, t.account_id, a.name AS account_name, t.from_account_id, t.to_account_id, t.exchange_rate::text, a.currency AS account_currency, fa.currency AS from_currency, ta.currency AS to_currency, t.status, t.created_at, t.updated_at, t.deleted_at`

	// TransactionPermissionSQL represents the complex logic to check if a user has access to a transaction.
	TransactionPermissionSQL = `(
		(t.type IN ('expense','income') AND EXISTS (SELECT 1 FROM user_accounts ua WHERE ua.user_id = $1 AND ua.account_id = t.account_id AND ua.status = 'active'))
		OR
		(t.type = 'transfer' AND EXISTS (SELECT 1 FROM user_accounts ua WHERE ua.user_id = $1 AND ua.account_id = t.from_account_id AND ua.status = 'active')
		                 AND EXISTS (SELECT 1 FROM user_accounts ua WHERE ua.user_id = $1 AND ua.account_id = t.to_account_id AND ua.status = 'active'))
	)`

	// ContactColumnsSQL represents the standard column selection for the contacts table.
	ContactColumnsSQL = `c.id, c.user_id, c.name, c.email, c.phone, c.avatar_url, c.linked_user_id, c.notes, c.created_at, c.updated_at, c.deleted_at, u.display_name AS linked_display_name, u.avatar_url AS linked_avatar_url`
)
