package profile

import (
	"context"
	"errors"
	"testing"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/domains/auth"
)

type fakeProfileRepo struct {
	userWithPass   *auth.UserWithPassword
	findByIDResult *auth.User
	updateResult   *auth.User
	avatarResult   *auth.User
	findByIDErr    error
	updateErr      error
	avatarErr      error
	updatePwdErr   error
}

func (r *fakeProfileRepo) CreateUser(context.Context, auth.UserWithPassword) error { return nil }
func (r *fakeProfileRepo) FindUserByID(context.Context, string) (*auth.User, error) {
	return r.findByIDResult, r.findByIDErr
}
func (r *fakeProfileRepo) FindUserByEmail(context.Context, string) (*auth.UserWithPassword, error) {
	return r.userWithPass, nil
}
func (r *fakeProfileRepo) FindUserByPhone(context.Context, string) (*auth.UserWithPassword, error) {
	return r.userWithPass, nil
}
func (r *fakeProfileRepo) FindUserByUsername(context.Context, string) (*auth.UserWithPassword, error) {
	return r.userWithPass, nil
}
func (r *fakeProfileRepo) UpdateUserProfile(context.Context, string, auth.UpdateProfileInput) (*auth.User, error) {
	return r.updateResult, r.updateErr
}
func (r *fakeProfileRepo) UpdateAvatarURL(context.Context, string, string) (*auth.User, error) {
	return r.avatarResult, r.avatarErr
}
func (r *fakeProfileRepo) UpdateUserSettings(context.Context, string, map[string]any) (*auth.User, error) {
	return nil, nil
}
func (r *fakeProfileRepo) UpdatePasswordHash(context.Context, string, string) error {
	return r.updatePwdErr
}

type fakeProfileHasher struct {
	compareErr error
	hashErr    error
}

func (h *fakeProfileHasher) Hash(string) (string, error) {
	if h.hashErr != nil {
		return "", h.hashErr
	}
	return "hashed", nil
}

func (h *fakeProfileHasher) Compare(string, string) error {
	return h.compareErr
}

type fakeAvatarStorage struct {
	url string
	err error
}

func (s *fakeAvatarStorage) UploadAvatar(context.Context, string, string, string, []byte) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.url, nil
}

func TestProfileServiceUploadAvatarSuccess(t *testing.T) {
	repo := &fakeProfileRepo{avatarResult: &auth.User{ID: "u1"}}
	hasher := &fakeProfileHasher{}
	storage := &fakeAvatarStorage{url: "https://cdn.example.com/goen/avatars/u1/file.jpg"}
	svc := NewService(repo, hasher, storage)

	updated, err := svc.UploadAvatar(context.Background(), "u1", "avatar.jpg", "image/jpeg", []byte("abc"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if updated == nil || updated.ID != "u1" {
		t.Fatalf("unexpected updated user: %#v", updated)
	}
}

func TestProfileServiceUploadAvatarStorageNotConfigured(t *testing.T) {
	repo := &fakeProfileRepo{}
	hasher := &fakeProfileHasher{}
	svc := NewService(repo, hasher, nil)

	_, err := svc.UploadAvatar(context.Background(), "u1", "avatar.jpg", "image/jpeg", []byte("abc"))
	if apperrors.KindOf(err) != apperrors.KindInternal {
		t.Fatalf("expected internal kind, got %v", apperrors.KindOf(err))
	}
}

func TestProfileServiceChangePasswordInvalidCurrentPassword(t *testing.T) {
	email := "u@example.com"
	repo := &fakeProfileRepo{
		findByIDResult: &auth.User{ID: "u1", Email: &email, Username: "u1"},
		userWithPass:   &auth.UserWithPassword{User: auth.User{ID: "u1"}, PasswordHash: "old"},
	}
	hasher := &fakeProfileHasher{compareErr: errors.New("invalid")}
	svc := NewService(repo, hasher, &fakeAvatarStorage{})

	err := svc.ChangePassword(context.Background(), "u1", "Wrong123", "NewPass123")
	if apperrors.KindOf(err) != apperrors.KindUnauth {
		t.Fatalf("expected unauthorized kind, got %v", apperrors.KindOf(err))
	}
}
