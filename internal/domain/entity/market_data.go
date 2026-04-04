package entity

import "time"

type SyncState struct {
	SyncKey            string     `json:"sync_key"`
	MinIntervalSeconds int        `json:"min_interval_seconds"`
	LastStartedAt      *time.Time `json:"last_started_at"`
	LastSuccessAt      *time.Time `json:"last_success_at"`
	LastFailureAt      *time.Time `json:"last_failure_at"`
	LastStatus         string     `json:"last_status"`
	LastError          *string    `json:"last_error"`
	NextDueAt          *time.Time `json:"next_due_at"`
	CooldownSeconds    int        `json:"cooldown_seconds"`
}

type RateLimit struct {
	PerMinute         int     `json:"per_minute"`
	PerHour           int     `json:"per_hour"`
	UsedMinute        int     `json:"used_minute"`
	UsedHour          int     `json:"used_hour"`
	RemainingMinute   int     `json:"remaining_minute"`
	RemainingHour     int     `json:"remaining_hour"`
	MinuteResetInSecs float64 `json:"minute_reset_in_seconds"`
	HourResetInSecs   float64 `json:"hour_reset_in_seconds"`
}
