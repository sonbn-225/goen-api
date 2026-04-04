package interfaces

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type ContactRepository interface {
	CreateContact(ctx context.Context, c entity.Contact) error
	GetContact(ctx context.Context, userID, contactID string) (*entity.Contact, error)
	ListContacts(ctx context.Context, userID string) ([]entity.Contact, error)
	UpdateContact(ctx context.Context, userID string, c entity.Contact) error
	DeleteContact(ctx context.Context, userID, contactID string) error

	// Finding users for linking
	FindUserByEmail(ctx context.Context, email string) (*entity.User, error)
	FindUserByPhone(ctx context.Context, phone string) (*entity.User, error)
}

type ContactService interface {
	Create(ctx context.Context, userID string, req dto.CreateContactRequest) (*dto.ContactResponse, error)
	Get(ctx context.Context, userID, contactID string) (*dto.ContactResponse, error)
	List(ctx context.Context, userID string) ([]dto.ContactResponse, error)
	Update(ctx context.Context, userID, contactID string, req dto.UpdateContactRequest) (*dto.ContactResponse, error)
	Delete(ctx context.Context, userID, contactID string) error
	GetOrCreateByName(ctx context.Context, userID, name string) (string, error)
}
