package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/domains/auth"
)

type fakeUserRepo struct {
	byID       map[string]*auth.UserWithPassword
	byEmail    map[string]*auth.UserWithPassword
	byPhone    map[string]*auth.UserWithPassword
	byUsername map[string]*auth.UserWithPassword
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		byID:       map[string]*auth.UserWithPassword{},
		byEmail:    map[string]*auth.UserWithPassword{},
		byPhone:    map[string]*auth.UserWithPassword{},
		byUsername: map[string]*auth.UserWithPassword{},
	}
}

func (r *fakeUserRepo) CreateUser(_ context.Context, user auth.UserWithPassword) error {
	u := user
	r.byID[u.ID] = &u
	if u.Email != nil {
		r.byEmail[*u.Email] = &u
	}
	if u.Phone != nil {
		r.byPhone[*u.Phone] = &u
	}
	r.byUsername[u.Username] = &u
	return nil
}

func (r *fakeUserRepo) FindUserByID(_ context.Context, userID string) (*auth.User, error) {
	u := r.byID[userID]
	if u == nil {
		return nil, nil
	}
	cloned := u.User
	return &cloned, nil
}

func (r *fakeUserRepo) FindUserByEmail(_ context.Context, email string) (*auth.UserWithPassword, error) {
	return r.byEmail[email], nil
}

func (r *fakeUserRepo) FindUserByPhone(_ context.Context, phone string) (*auth.UserWithPassword, error) {
	return r.byPhone[phone], nil
}

func (r *fakeUserRepo) FindUserByUsername(_ context.Context, username string) (*auth.UserWithPassword, error) {
	return r.byUsername[username], nil
}

func (r *fakeUserRepo) UpdateUserProfile(_ context.Context, userID string, input auth.UpdateProfileInput) (*auth.User, error) {
	u := r.byID[userID]
	if u == nil {
		return nil, nil
	}
	if input.DisplayName != nil {
		u.DisplayName = input.DisplayName
	}
	if input.Email != nil {
		u.Email = input.Email
	}
	if input.Phone != nil {
		u.Phone = input.Phone
	}
	if input.Username != nil {
		u.Username = *input.Username
	}
	u.UpdatedAt = time.Now().UTC()
	cloned := u.User
	return &cloned, nil
}

func (r *fakeUserRepo) UpdateUserSettings(_ context.Context, userID string, patch map[string]any) (*auth.User, error) {
	u := r.byID[userID]
	if u == nil {
		return nil, nil
	}
	if u.Settings == nil {
		u.Settings = map[string]any{}
	}
	for k, v := range patch {
		u.Settings[k] = v
	}
	u.UpdatedAt = time.Now().UTC()
	cloned := u.User
	return &cloned, nil
}

func (r *fakeUserRepo) UpdateAvatarURL(_ context.Context, userID, avatarURL string) (*auth.User, error) {
	u := r.byID[userID]
	if u == nil {
		return nil, nil
	}
	u.AvatarURL = strPtr(avatarURL)
	u.UpdatedAt = time.Now().UTC()
	cloned := u.User
	return &cloned, nil
}

func (r *fakeUserRepo) UpdatePasswordHash(_ context.Context, userID, passwordHash string) error {
	u := r.byID[userID]
	if u == nil {
		return nil
	}
	u.PasswordHash = passwordHash
	u.UpdatedAt = time.Now().UTC()
	return nil
}

type fakeHasher struct{}

func (h fakeHasher) Hash(password string) (string, error) {
	return "hashed-" + password, nil
}

func (h fakeHasher) Compare(hash, password string) error {
	if hash != "hashed-"+password {
		return errors.New("invalid credentials")
	}
	return nil
}

type fakeIssuer struct{}

func (i fakeIssuer) Issue(userID string) (string, error) {
	return "token-" + userID, nil
}

func TestServiceSignupSuccess(t *testing.T) {
	repo := newFakeUserRepo()
	svc := auth.NewService(repo, fakeHasher{}, fakeIssuer{}, 60)

	result, err := svc.Signup(context.Background(), auth.SignupRequest{
		Email:    "User@Example.com",
		Username: "User",
		Password: "Password123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if result.User.Username != "user" {
		t.Fatalf("expected normalized username user, got %s", result.User.Username)
	}
	if result.User.Email == nil || *result.User.Email != "user@example.com" {
		t.Fatalf("expected normalized email user@example.com, got %#v", result.User.Email)
	}

	stored := repo.byID[result.User.ID]
	if stored == nil {
		t.Fatal("expected created user in repo")
	}
	if stored.PasswordHash != "hashed-Password123" {
		t.Fatalf("expected stored hash hashed-Password123, got %s", stored.PasswordHash)
	}
}

func TestServiceSigninSuccess(t *testing.T) {
	repo := newFakeUserRepo()
	_ = repo.CreateUser(context.Background(), auth.UserWithPassword{
		User: auth.User{
			ID:       "u1",
			Username: "user",
			Email:    strPtr("user@example.com"),
		},
		PasswordHash: "hashed-Password123",
	})

	svc := auth.NewService(repo, fakeHasher{}, fakeIssuer{}, 60)
	result, err := svc.Signin(context.Background(), auth.SigninRequest{Login: "USER@example.com", Password: "Password123"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.User.ID != "u1" {
		t.Fatalf("expected user id u1, got %s", result.User.ID)
	}
	if result.AccessToken != "token-u1" {
		t.Fatalf("expected token token-u1, got %s", result.AccessToken)
	}
}

func TestServiceRefreshSuccess(t *testing.T) {
	repo := newFakeUserRepo()
	_ = repo.CreateUser(context.Background(), auth.UserWithPassword{
		User: auth.User{
			ID:       "u-refresh",
			Username: "u_refresh",
			Email:    strPtr("refresh@example.com"),
		},
		PasswordHash: "hashed-Password123",
	})

	svc := auth.NewService(repo, fakeHasher{}, fakeIssuer{}, 60)
	result, err := svc.Refresh(context.Background(), "u-refresh")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.AccessToken != "token-u-refresh" {
		t.Fatalf("expected token token-u-refresh, got %s", result.AccessToken)
	}
	if result.User.ID != "u-refresh" {
		t.Fatalf("expected user id u-refresh, got %s", result.User.ID)
	}
}

func TestServiceChangePasswordSuccess(t *testing.T) {
	repo := newFakeUserRepo()
	_ = repo.CreateUser(context.Background(), auth.UserWithPassword{
		User: auth.User{
			ID:       "u-password",
			Username: "u_password",
			Email:    strPtr("password@example.com"),
		},
		PasswordHash: "hashed-Password123",
	})

	svc := auth.NewService(repo, fakeHasher{}, fakeIssuer{}, 60)
	err := svc.ChangePassword(context.Background(), "u-password", "Password123", "NewPassword9")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	stored := repo.byID["u-password"]
	if stored == nil {
		t.Fatal("expected user in repo")
	}
	if stored.PasswordHash != "hashed-NewPassword9" {
		t.Fatalf("expected updated hash hashed-NewPassword9, got %s", stored.PasswordHash)
	}
}

func TestServiceSignupValidationRequiresEmailOrPhone(t *testing.T) {
	repo := newFakeUserRepo()
	svc := auth.NewService(repo, fakeHasher{}, fakeIssuer{}, 60)

	_, err := svc.Signup(context.Background(), auth.SignupRequest{
		Username: "user",
		Password: "Password123",
	})
	assertErrKind(t, err, apperrors.KindValidation)
}

func TestServiceSignupValidationShortPassword(t *testing.T) {
	repo := newFakeUserRepo()
	svc := auth.NewService(repo, fakeHasher{}, fakeIssuer{}, 60)

	_, err := svc.Signup(context.Background(), auth.SignupRequest{
		Email:    "user@example.com",
		Username: "user",
		Password: "Pass1",
	})
	assertErrKind(t, err, apperrors.KindValidation)
}

func TestServiceSignupConflictExistingEmail(t *testing.T) {
	repo := newFakeUserRepo()
	_ = repo.CreateUser(context.Background(), auth.UserWithPassword{
		User: auth.User{
			ID:       "u-existing-email",
			Username: "existing_user",
			Email:    strPtr("user@example.com"),
		},
		PasswordHash: "hashed-Password123",
	})

	svc := auth.NewService(repo, fakeHasher{}, fakeIssuer{}, 60)
	_, err := svc.Signup(context.Background(), auth.SignupRequest{
		Email:    "user@example.com",
		Username: "another_user",
		Password: "Password123",
	})
	assertErrKind(t, err, apperrors.KindConflict)
}

func TestServiceSignupConflictExistingUsername(t *testing.T) {
	repo := newFakeUserRepo()
	_ = repo.CreateUser(context.Background(), auth.UserWithPassword{
		User: auth.User{
			ID:       "u-existing-username",
			Username: "existing_user",
			Email:    strPtr("existing@example.com"),
		},
		PasswordHash: "hashed-Password123",
	})

	svc := auth.NewService(repo, fakeHasher{}, fakeIssuer{}, 60)
	_, err := svc.Signup(context.Background(), auth.SignupRequest{
		Email:    "new@example.com",
		Username: "existing_user",
		Password: "Password123",
	})
	assertErrKind(t, err, apperrors.KindConflict)
}

func TestServiceSigninUnauthorizedWrongPassword(t *testing.T) {
	repo := newFakeUserRepo()
	_ = repo.CreateUser(context.Background(), auth.UserWithPassword{
		User: auth.User{
			ID:       "u-signin",
			Username: "signin_user",
			Email:    strPtr("signin@example.com"),
		},
		PasswordHash: "hashed-Password123",
	})

	svc := auth.NewService(repo, fakeHasher{}, fakeIssuer{}, 60)
	_, err := svc.Signin(context.Background(), auth.SigninRequest{Login: "signin@example.com", Password: "WrongPassword9"})
	assertErrKind(t, err, apperrors.KindUnauth)
}

func TestServiceSigninUnauthorizedUserNotFound(t *testing.T) {
	repo := newFakeUserRepo()
	svc := auth.NewService(repo, fakeHasher{}, fakeIssuer{}, 60)

	_, err := svc.Signin(context.Background(), auth.SigninRequest{Login: "missing@example.com", Password: "Password123"})
	assertErrKind(t, err, apperrors.KindUnauth)
}

func TestServiceRefreshUnauthorizedUserNotFound(t *testing.T) {
	repo := newFakeUserRepo()
	svc := auth.NewService(repo, fakeHasher{}, fakeIssuer{}, 60)

	_, err := svc.Refresh(context.Background(), "missing-user")
	assertErrKind(t, err, apperrors.KindUnauth)
}

func TestServiceUpdateMyAvatarSuccess(t *testing.T) {
	repo := newFakeUserRepo()
	_ = repo.CreateUser(context.Background(), auth.UserWithPassword{
		User: auth.User{
			ID:       "u-avatar",
			Username: "avatar_user",
			Email:    strPtr("avatar@example.com"),
		},
		PasswordHash: "hashed-Password123",
	})

	svc := auth.NewService(repo, fakeHasher{}, fakeIssuer{}, 60)
	updated, err := svc.UpdateMyAvatar(context.Background(), "u-avatar", "/uploads/avatars/file.jpg")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.AvatarURL == nil || *updated.AvatarURL != "/uploads/avatars/file.jpg" {
		t.Fatalf("expected updated avatar_url, got %#v", updated.AvatarURL)
	}
}

func TestServiceUpdateMyAvatarNotFound(t *testing.T) {
	repo := newFakeUserRepo()
	svc := auth.NewService(repo, fakeHasher{}, fakeIssuer{}, 60)

	_, err := svc.UpdateMyAvatar(context.Background(), "missing-user", "/uploads/avatars/file.jpg")
	assertErrKind(t, err, apperrors.KindNotFound)
}

func TestServiceChangePasswordValidationWeakPassword(t *testing.T) {
	repo := newFakeUserRepo()
	_ = repo.CreateUser(context.Background(), auth.UserWithPassword{
		User: auth.User{
			ID:       "u-weak-password",
			Username: "weak_user",
			Email:    strPtr("weak@example.com"),
		},
		PasswordHash: "hashed-Password123",
	})

	svc := auth.NewService(repo, fakeHasher{}, fakeIssuer{}, 60)
	err := svc.ChangePassword(context.Background(), "u-weak-password", "Password123", "alllowercase")
	assertErrKind(t, err, apperrors.KindValidation)
}

func TestServiceChangePasswordUnauthorizedInvalidCurrentPassword(t *testing.T) {
	repo := newFakeUserRepo()
	_ = repo.CreateUser(context.Background(), auth.UserWithPassword{
		User: auth.User{
			ID:       "u-wrong-current",
			Username: "wrong_current_user",
			Email:    strPtr("wrong-current@example.com"),
		},
		PasswordHash: "hashed-Password123",
	})

	svc := auth.NewService(repo, fakeHasher{}, fakeIssuer{}, 60)
	err := svc.ChangePassword(context.Background(), "u-wrong-current", "WrongPassword9", "NewPassword9")
	assertErrKind(t, err, apperrors.KindUnauth)
}

func assertErrKind(t *testing.T, err error, expected apperrors.Kind) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error kind %s, got nil", expected)
	}
	if got := apperrors.KindOf(err); got != expected {
		t.Fatalf("expected error kind %s, got %s (err=%v)", expected, got, err)
	}
}

func strPtr(v string) *string {
	return &v
}
