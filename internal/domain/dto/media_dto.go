package dto

import "time"

// MediaResponse represents the metadata of a media file (image, document, etc.) stored in the system.
// Used in: MediaHandler, MediaService, MediaInterface
// MediaResponse represents the metadata of a media file (image, document, etc.) stored in the system.
// Used in: MediaHandler, MediaService, MediaInterface
type MediaResponse struct {
	Key       string    `json:"key"`        // Unique storage key (e.g., S3 path)
	URL       string    `json:"url"`        // Presigned or public access URL
	Size      int64     `json:"size"`       // File size in bytes
	UpdatedAt time.Time `json:"updated_at"` // Last modified timestamp
}
