package entity

import (
	"time"
)

type Tag struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	NameVI    *string   `json:"name_vi,omitempty"`
	NameEN    *string   `json:"name_en,omitempty"`
	Color     *string   `json:"color,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
