package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type ContactRepository interface {
	CreateContact(ctx context.Context, c entity.Contact) error
	GetContact(ctx context.Context, userID, contactID uuid.UUID) (*entity.Contact, error)
	ListContacts(ctx context.Context, userID uuid.UUID) ([]entity.Contact, error)
	UpdateContact(ctx context.Context, userID uuid.UUID, c entity.Contact) error
	DeleteContact(ctx context.Context, userID, contactID uuid.UUID) error

	// Finding users for linking
	FindUserByEmail(ctx context.Context, email string) (*entity.User, error)
	FindUserByPhone(ctx context.Context, phone string) (*entity.User, error)
}

type ContactService interface {
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateContactRequest) (*dto.ContactResponse, error)
	Get(ctx context.Context, userID, contactID uuid.UUID) (*dto.ContactResponse, error)
	List(ctx context.Context, userID uuid.UUID) ([]dto.ContactResponse, error)
	Update(ctx context.Context, userID, contactID uuid.UUID, req dto.UpdateContactRequest) (*dto.ContactResponse, error)
	Delete(ctx context.Context, userID, contactID uuid.UUID) error
	GetOrCreateByName(ctx context.Context, userID uuid.UUID, name string) (uuid.UUID, error)
}

