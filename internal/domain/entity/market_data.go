package entity

import "time"

// SyncState tracks the synchronization status of external market data (e.g. security prices).
type SyncState struct {
	SyncKey            string     `json:"sync_key"`               // Unique key identifying the sync task (e.g., "vn_stock_prices")
	MinIntervalSeconds int        `json:"min_interval_seconds"`  // Minimum time to wait between sync attempts
	LastStartedAt      *time.Time `json:"last_started_at"`      // Timestamp of the last started sync attempt
	LastSuccessAt      *time.Time `json:"last_success_at"`      // Timestamp of the last successful sync completion
	LastFailureAt      *time.Time `json:"last_failure_at"`      // Timestamp of the last failed sync attempt
	LastStatus         string     `json:"last_status"`          // Status of the last attempt (success/failure)
	LastError          *string    `json:"last_error"`           // Error message from the last failure
	NextDueAt          *time.Time `json:"next_due_at"`          // Timestamp when the next sync is scheduled
	CooldownSeconds    int        `json:"cooldown_seconds"`     // Dynamic cooldown period after a failure
}

// RateLimit represents the API consumption limits for an external data provider.
type RateLimit struct {
	PerMinute         int     `json:"per_minute"`           // Maximum allowed requests per minute
	PerHour           int     `json:"per_hour"`             // Maximum allowed requests per hour
	UsedMinute        int     `json:"used_minute"`          // Number of requests already used this minute
	UsedHour          int     `json:"used_hour"`            // Number of requests already used this hour
	RemainingMinute   int     `json:"remaining_minute"`     // Requests remaining for the current minute
	RemainingHour     int     `json:"remaining_hour"`       // Requests remaining for the current hour
	MinuteResetInSecs float64 `json:"minute_reset_in_seconds"` // Seconds until the minute limit resets
	HourResetInSecs   float64 `json:"hour_reset_in_seconds"`   // Seconds until the hour limit resets
}
