package dto

import "time"

type MediaResponse struct {
	Key       string    `json:"key"`
	URL       string    `json:"url"`
	Size      int64     `json:"size"`
	UpdatedAt time.Time `json:"updated_at"`
}
