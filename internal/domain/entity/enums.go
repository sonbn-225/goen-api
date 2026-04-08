package entity

// AccountType defines the category of a financial account.
type AccountType string

const (
	AccountTypeBank    AccountType = "bank"    // Regular bank account
	AccountTypeWallet  AccountType = "wallet"  // Digital wallet (e.g. MoMo, Zalopay)
	AccountTypeCash    AccountType = "cash"    // Physical cash on hand
	AccountTypeBroker  AccountType = "broker"  // Investment/Stock brokerage account
	AccountTypeCard    AccountType = "card"    // Credit or debit card
	AccountTypeSavings AccountType = "savings" // Dedicated savings account or goal
)

// AccountStatus defines whether an account is currently active or has been closed.
type AccountStatus string

const (
	AccountStatusActive AccountStatus = "active"
	AccountStatusClosed AccountStatus = "closed"
)

// AccountSharePermission defines the level of access granted to a shared account.
type AccountSharePermission string

const (
	AccountSharePermissionOwner  AccountSharePermission = "owner"  // Full control (only for the creator)
	AccountSharePermissionViewer AccountSharePermission = "viewer" // Read-only access
	AccountSharePermissionEditor AccountSharePermission = "editor" // Read and write access
)

// AccountShareStatus defines the current state of an account sharing relationship.
type AccountShareStatus string

const (
	AccountShareStatusActive  AccountShareStatus = "active"
	AccountShareStatusRevoked AccountShareStatus = "revoked"
)

// BudgetPeriod defines the recurring interval for a budget.
type BudgetPeriod string

const (
	BudgetPeriodMonth  BudgetPeriod = "month"
	BudgetPeriodWeek   BudgetPeriod = "week"
	BudgetPeriodCustom BudgetPeriod = "custom"
)

// BudgetRolloverMode defines how unused budget amounts are handled at the end of a period.
type BudgetRolloverMode string

const (
	BudgetRolloverModeNone        BudgetRolloverMode = "none"         // No rollover
	BudgetRolloverModeAddPositive BudgetRolloverMode = "add_positive" // Only add remaining positive amounts to next period
	BudgetRolloverModeAddAll      BudgetRolloverMode = "add_all"      // Add both remaining positive and negative balances
)

// DebtDirection indicates if the user is the lender or the borrower.
type DebtDirection string

const (
	DebtDirectionLent     DebtDirection = "lent"     // User lent money to someone else
	DebtDirectionBorrowed DebtDirection = "borrowed" // User borrowed money from someone else
)

// DebtStatus defines the current state of a debt or loan record.
type DebtStatus string

const (
	DebtStatusActive    DebtStatus = "active"
	DebtStatusPaid      DebtStatus = "paid"
	DebtStatusCancelled DebtStatus = "cancelled"
)

// ImportMappingRuleKind defines the type of entity being mapped during transaction import.
type ImportMappingRuleKind string

const (
	ImportMappingRuleKindAccount  ImportMappingRuleKind = "account"
	ImportMappingRuleKindCategory ImportMappingRuleKind = "category"
)

// SecurityEventType defines common corporate actions that affect security holdings.
type SecurityEventType string

const (
	SecurityEventTypeCashDividend  SecurityEventType = "cash_dividend"
	SecurityEventTypeStockDividend SecurityEventType = "stock_dividend"
	SecurityEventTypeSplit         SecurityEventType = "split"
	SecurityEventTypeMerger        SecurityEventType = "merger"
)

// SecurityEventElectionStatus defines the state of a user's claim on a corporate action.
type SecurityEventElectionStatus string

const (
	SecurityEventElectionStatusEligible  SecurityEventElectionStatus = "eligible"  // User is entitled to the action but hasn't claimed it
	SecurityEventElectionStatusClaimed   SecurityEventElectionStatus = "claimed"   // User has successfully claimed the entitlement
	SecurityEventElectionStatusDismissed SecurityEventElectionStatus = "dismissed" // User has opted out of the action
)

// TradeSide indicates whether a security trade is a purchase or a sale.
type TradeSide string

const (
	TradeSideBuy  TradeSide = "buy"
	TradeSideSell TradeSide = "sell"
)

// ShareLotStatus defines whether a lot of purchased shares is still held or fully sold.
type ShareLotStatus string

const (
	ShareLotStatusActive ShareLotStatus = "active" // Some or all shares in this lot are still held
	ShareLotStatusClosed ShareLotStatus = "closed" // All shares in this lot have been sold
)

// RotatingSavingsCycleFrequency defines how often members contribute to a rotating savings group.
type RotatingSavingsCycleFrequency string

const (
	RotatingSavingsCycleFrequencyWeekly  RotatingSavingsCycleFrequency = "weekly"
	RotatingSavingsCycleFrequencyMonthly RotatingSavingsCycleFrequency = "monthly"
)

// RotatingSavingsStatus defines the lifecycle stage of a rotating savings group.
type RotatingSavingsStatus string

const (
	RotatingSavingsStatusActive    RotatingSavingsStatus = "active"
	RotatingSavingsStatusCompleted RotatingSavingsStatus = "completed"
	RotatingSavingsStatusClosed    RotatingSavingsStatus = "closed"
)

// RotatingSavingsContributionKind defines the type of payment within a rotating savings cycle.
type RotatingSavingsContributionKind string

const (
	RotatingSavingsContributionKindContribution RotatingSavingsContributionKind = "contribution" // Regular member payment
	RotatingSavingsContributionKindPayout       RotatingSavingsContributionKind = "payout"       // Member receiving the cycle's total pot
	RotatingSavingsContributionKindCollected    RotatingSavingsContributionKind = "collected"    // Pot already received by someone else
)

// RotatingSavingsAuditAction defines the types of recorded events for auditing rotating savings groups.
type RotatingSavingsAuditAction string

const (
	RotatingSavingsAuditActionGroupCreated        RotatingSavingsAuditAction = "group_created"
	RotatingSavingsAuditActionGroupUpdated        RotatingSavingsAuditAction = "group_updated"
	RotatingSavingsAuditActionContributionCreated RotatingSavingsAuditAction = "contribution_created"
)

// SavingsStatus defines the current state of a dedicated savings product or goal.
type SavingsStatus string

const (
	SavingsStatusActive  SavingsStatus = "active"
	SavingsStatusMatured SavingsStatus = "matured" // Savings term has reached maturity
	SavingsStatusClosed  SavingsStatus = "closed"  // Savings goal has been withdrawn or cancelled
)

// TransactionType defines the fundamental category of a financial transaction.
type TransactionType string

const (
	TransactionTypeExpense  TransactionType = "expense"
	TransactionTypeIncome   TransactionType = "income"
	TransactionTypeTransfer TransactionType = "transfer"
)

// TransactionStatus defines the processing state of a transaction.
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"   // Transaction is planned but hasn't occurred yet
	TransactionStatusPosted    TransactionStatus = "posted"    // Transaction has been finalized and recorded
	TransactionStatusCancelled TransactionStatus = "cancelled" // Transaction was voided
)

// CategoryType defines the fundamental category of a transaction (income or expense).
type CategoryType string

const (
	CategoryTypeExpense CategoryType = "expense"
	CategoryTypeIncome  CategoryType = "income"
)
