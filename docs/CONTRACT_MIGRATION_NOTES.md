# Contract Migration Notes

## Scope

- Development phase keeps API namespace at `/api/v1` only.
- Auth, Account, and Transaction modules are now persisted in PostgreSQL.
- No migration changes are required for this step; existing SQL migrations are reused.

## Route Mapping (Draft v2 -> Current v1)

| Draft path | Current path | Status |
|---|---|---|
| `POST /api/v2/auth/register` | `POST /api/v1/auth/register` | moved to v1 |
| `POST /api/v2/auth/login` | `POST /api/v1/auth/login` | moved to v1 |
| `POST /api/v2/auth/signup` | `POST /api/v1/auth/signup` | moved to v1 |
| `POST /api/v2/auth/signin` | `POST /api/v1/auth/signin` | moved to v1 |
| `POST /api/v2/auth/refresh` | `POST /api/v1/auth/refresh` | moved to v1 |
| `GET /api/v2/auth/me` | `GET /api/v1/auth/me` | moved to v1 |
| `PATCH /api/v2/auth/me/profile` | `PATCH /api/v1/auth/me/profile` | moved to v1 |
| `PATCH /api/v2/auth/me/settings` | `PATCH /api/v1/auth/me/settings` | moved to v1 |
| `POST /api/v2/auth/me/change-password` | `POST /api/v1/auth/me/change-password` | moved to v1 |
| `GET /api/v2/accounts` | `GET /api/v1/accounts` | moved to v1 |
| `POST /api/v2/accounts` | `POST /api/v1/accounts` | moved to v1 |
| `GET /api/v2/transactions` | `GET /api/v1/transactions` | moved to v1 |
| `POST /api/v2/transactions` | `POST /api/v1/transactions` | moved to v1 |
| `PATCH /api/v2/transactions/{transactionId}` | `PATCH /api/v1/transactions/{transactionId}` | moved to v1 |
| `GET /api/v2/transactions/{transactionId}/group-expense-participants` | `GET /api/v1/transactions/{transactionId}/group-expense-participants` | moved to v1 |
| `GET /api/v2/investment-accounts` | `GET /api/v1/investment-accounts` | moved to v1 |
| `GET /api/v2/investment-accounts/{investmentAccountId}` | `GET /api/v1/investment-accounts/{investmentAccountId}` | moved to v1 |
| `PATCH /api/v2/investment-accounts/{investmentAccountId}` | `PATCH /api/v1/investment-accounts/{investmentAccountId}` | moved to v1 |
| `POST /api/v2/investment-accounts/{investmentAccountId}/trades` | `POST /api/v1/investment-accounts/{investmentAccountId}/trades` | moved to v1 |
| `GET /api/v2/investment-accounts/{investmentAccountId}/trades` | `GET /api/v1/investment-accounts/{investmentAccountId}/trades` | moved to v1 |
| `GET /api/v2/investment-accounts/{investmentAccountId}/holdings` | `GET /api/v1/investment-accounts/{investmentAccountId}/holdings` | moved to v1 |
| `GET /api/v2/securities` | `GET /api/v1/securities` | moved to v1 |
| `GET /api/v2/securities/{securityId}` | `GET /api/v1/securities/{securityId}` | moved to v1 |
| `GET /api/v2/securities/{securityId}/prices-daily` | `GET /api/v1/securities/{securityId}/prices-daily` | moved to v1 |
| `GET /api/v2/securities/{securityId}/events` | `GET /api/v1/securities/{securityId}/events` | moved to v1 |
| `GET /api/v2/savings/instruments` | `GET /api/v1/savings/instruments` | moved to v1 |
| `POST /api/v2/savings/instruments` | `POST /api/v1/savings/instruments` | moved to v1 |
| `GET /api/v2/savings/instruments/{instrumentId}` | `GET /api/v1/savings/instruments/{instrumentId}` | moved to v1 |
| `PATCH /api/v2/savings/instruments/{instrumentId}` | `PATCH /api/v1/savings/instruments/{instrumentId}` | moved to v1 |
| `DELETE /api/v2/savings/instruments/{instrumentId}` | `DELETE /api/v1/savings/instruments/{instrumentId}` | moved to v1 |
| `GET /api/v2/rotating-savings/groups` | `GET /api/v1/rotating-savings/groups` | moved to v1 |
| `POST /api/v2/rotating-savings/groups` | `POST /api/v1/rotating-savings/groups` | moved to v1 |
| `GET /api/v2/rotating-savings/groups/{groupId}` | `GET /api/v1/rotating-savings/groups/{groupId}` | moved to v1 |
| `PATCH /api/v2/rotating-savings/groups/{groupId}` | `PATCH /api/v1/rotating-savings/groups/{groupId}` | moved to v1 |
| `DELETE /api/v2/rotating-savings/groups/{groupId}` | `DELETE /api/v1/rotating-savings/groups/{groupId}` | moved to v1 |
| `GET /api/v2/rotating-savings/groups/{groupId}/contributions` | `GET /api/v1/rotating-savings/groups/{groupId}/contributions` | moved to v1 |
| `POST /api/v2/rotating-savings/groups/{groupId}/contributions` | `POST /api/v1/rotating-savings/groups/{groupId}/contributions` | moved to v1 |
| `DELETE /api/v2/rotating-savings/groups/{groupId}/contributions/{contributionId}` | `DELETE /api/v1/rotating-savings/groups/{groupId}/contributions/{contributionId}` | moved to v1 |
| `GET /api/v2/reports/dashboard` | `GET /api/v1/reports/dashboard` | moved to v1 |

## Break Changes and Compatibility Notes

1. API prefix is fixed to `/api/v1` for now.
2. Auth data is persisted in `users` table instead of in-memory process state.
3. Account creation now writes both `accounts` and `user_accounts` rows for ownership.
4. Transaction creation now validates account ownership using `user_accounts` before insert.
5. `accounts.balance` is accepted by API payload for compatibility, but current persistence model stores source-of-truth as transactions.
6. Transaction API supports `expense`, `income`, and `transfer`.
7. Group expense is implemented as a transaction extension (not a standalone domain module) and uses extension fields on `POST /transactions`.
8. Investment domain endpoints are now active in v1 for investment accounts, trades, holdings, and securities read APIs.
9. Savings instrument endpoints are active in v1 and support auto-linked savings account creation from a parent bank/wallet account.
10. Rotating savings endpoints are active in v1 with group/contribution/audit schedule responses and linked transaction writes.

## Investment Contract (v1)

### Investment Accounts

1. `GET /api/v1/investment-accounts`: list accessible investment accounts.
2. `GET /api/v1/investment-accounts/{investmentAccountId}`: get details by ID.
3. `PATCH /api/v1/investment-accounts/{investmentAccountId}`: patch fee/tax settings (owner/editor permission required).

### Trades and Holdings

1. `POST /api/v1/investment-accounts/{investmentAccountId}/trades`: create trade (`buy`/`sell`) and refresh holding snapshot.
2. `GET /api/v1/investment-accounts/{investmentAccountId}/trades`: list trades.
3. `GET /api/v1/investment-accounts/{investmentAccountId}/holdings`: list holdings.
4. Trade validation enforces decimal `quantity`/`price` > 0 and non-negative `fees`/`taxes`.

### Securities and Market Data Reads

1. `GET /api/v1/securities`: list securities.
2. `GET /api/v1/securities/{securityId}`: get security details.
3. `GET /api/v1/securities/{securityId}/prices-daily?from=YYYY-MM-DD&to=YYYY-MM-DD`: list price dailies.
4. `GET /api/v1/securities/{securityId}/events?from=YYYY-MM-DD&to=YYYY-MM-DD`: list security events.

## Report Contract (v1)

1. `GET /api/v1/reports/dashboard`: return dashboard summary payload:
	- `total_balances`
	- `cashflow_6_months`
	- `top_expenses_month`
2. Endpoint requires authenticated user context.

## Savings Contract (v1)

1. `GET /api/v1/savings/instruments`: list user-accessible savings instruments.
2. `POST /api/v1/savings/instruments`: create a savings instrument.
3. `GET /api/v1/savings/instruments/{instrumentId}`: get savings instrument details.
4. `PATCH /api/v1/savings/instruments/{instrumentId}`: patch mutable fields.
5. `DELETE /api/v1/savings/instruments/{instrumentId}`: delete savings instrument.
6. Create rule: require either `savings_account_id` or `parent_account_id`.
7. When only `parent_account_id` is provided, backend auto-creates a linked `savings` account and writes an initial transfer transaction from parent to savings account.

## Rotating Savings Contract (v1)

1. `GET /api/v1/rotating-savings/groups`: list rotating savings groups with summary stats.
2. `POST /api/v1/rotating-savings/groups`: create rotating savings group.
3. `GET /api/v1/rotating-savings/groups/{groupId}`: get group detail with schedule, contributions, and audit logs.
4. `PATCH /api/v1/rotating-savings/groups/{groupId}`: patch group metadata/status.
5. `DELETE /api/v1/rotating-savings/groups/{groupId}`: delete group and cleanup linked records.
6. `GET /api/v1/rotating-savings/groups/{groupId}/contributions`: list contributions/payouts in group.
7. `POST /api/v1/rotating-savings/groups/{groupId}/contributions`: create contribution/payout and linked transaction entry.
8. `DELETE /api/v1/rotating-savings/groups/{groupId}/contributions/{contributionId}`: delete contribution and cleanup linked transaction entry.

## Transaction Transfer Contract (v1)

### Request

- Endpoint: `POST /api/v1/transactions`
- Rule: when `type=transfer`, `account_id` must be omitted and both `from_account_id`/`to_account_id` are required.

```json
{
	"type": "transfer",
	"from_account_id": "acc_source_001",
	"to_account_id": "acc_target_002",
	"amount": "250000.00",
	"note": "Move cash to savings"
}
```

### Success Response

```json
{
	"data": {
		"id": "tx_01",
		"user_id": "user_01",
		"from_account_id": "acc_source_001",
		"to_account_id": "acc_target_002",
		"type": "transfer",
		"amount": "250000",
		"note": "Move cash to savings",
		"created_at": "2026-04-03T10:00:00Z"
	}
}
```

### Validation Rules

1. `type` must be one of: `expense`, `income`, `transfer`.
2. `amount` must be greater than `0`.
3. `transfer` requires both `from_account_id` and `to_account_id`.
4. `from_account_id` and `to_account_id` must be different.
5. `expense`/`income` require `account_id` and must not include `from_account_id`/`to_account_id`.
6. `line_items` is required for non-transfer transactions.
7. `line_items` must be empty for transfer transactions.
8. Each `line_items[]` entry requires `category_id` and `amount > 0`.

## Transaction Line Items (v1)

### Create Contract

1. `POST /api/v1/transactions` accepts `line_items[]` payload.
2. `line_items[].tag_ids` is optional and validated against user-owned tags.
3. If top-level `note` is provided and first line item note is empty, service copies note into first line item note.

### Update Contract

1. `PATCH /api/v1/transactions/{transactionId}` is active with `UpdateTransactionRequest`.
2. `line_items` supports full replacement semantics for non-transfer transactions.
3. For transfer transactions, `line_items` must remain empty.
4. `group_participants` supports replacement of unsettled participants for expense transactions.

### Persistence

1. Transaction row, `transaction_line_items`, and `transaction_line_item_tags` are written atomically in one DB transaction.
2. Category existence and active-state checks are enforced before line-item insert.

## Transaction Group Expense Extension (v1)

### Request Fields (Optional, `type=expense` only)

1. `owner_original_amount`: optional decimal string `>= 0`.
2. `group_participants`: optional array.
3. `group_participants[].participant_name`: required non-empty string.
4. `group_participants[].original_amount`: required decimal string `> 0`.
5. `group_participants[].share_amount`: optional decimal string `> 0`.

### Behavior Rules

1. Extension fields are rejected for non-`expense` transaction types.
2. If `share_amount` is omitted, backend auto-calculates proportional shares using:
	- numerator: each participant `original_amount`
	- denominator: `owner_original_amount + sum(participant.original_amount)`
	- base amount: transaction `amount`
3. Transaction row and `group_expense_participants` rows are persisted atomically.
4. During transaction PATCH, unsettled participants are replaced while settled participants are preserved.

## Layer Responsibility Clarification

- `internal/infra`: external system connectivity (for example Postgres pool, migration runner, Redis client).
- `internal/repository`: repository implementations with query and persistence logic.
- `internal/core/security`: reusable security primitives (JWT issue/verify helpers, password hashing) used by use cases and middleware.

## Next Migration Tasks

1. Add `user_accounts` and transaction integration tests with a real Postgres test container.
2. Add compatibility test matrix for all v1 routes before introducing any new namespace.
