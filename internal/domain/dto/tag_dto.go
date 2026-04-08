package dto
 
import (
	"github.com/google/uuid"
)
 
// CreateTagRequest is the payload for creating a new tag.
// Used in: TagHandler, TagService, TagInterface
type CreateTagRequest struct {
	NameVI *string `json:"name_vi,omitempty"` // Name of the tag in Vietnamese
	NameEN *string `json:"name_en,omitempty"` // Name of the tag in English
	Color  *string `json:"color,omitempty"`   // UI color representation in hex
}
 
// TagResponse represents a tag's information.
// Used in: TagHandler, TagService, TagInterface
type TagResponse struct {
	ID     uuid.UUID `json:"id"`             // Unique tag identifier
	UserID uuid.UUID `json:"user_id"`        // ID of the user who owns this tag
	NameVI *string   `json:"name_vi,omitempty"` // Name in Vietnamese
	NameEN *string   `json:"name_en,omitempty"` // Name in English
	Color  *string   `json:"color,omitempty"`   // UI color in hex
}
