package setting

import (
	"context"
	"errors"
	"testing"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/domains/auth"
)

type fakeSettingRepo struct {
	updated *auth.User
	err     error
}

func (r *fakeSettingRepo) CreateUser(context.Context, auth.UserWithPassword) error  { return nil }
func (r *fakeSettingRepo) FindUserByID(context.Context, string) (*auth.User, error) { return nil, nil }
func (r *fakeSettingRepo) FindUserByEmail(context.Context, string) (*auth.UserWithPassword, error) {
	return nil, nil
}
func (r *fakeSettingRepo) FindUserByPhone(context.Context, string) (*auth.UserWithPassword, error) {
	return nil, nil
}
func (r *fakeSettingRepo) FindUserByUsername(context.Context, string) (*auth.UserWithPassword, error) {
	return nil, nil
}
func (r *fakeSettingRepo) UpdateUserProfile(context.Context, string, auth.UpdateProfileInput) (*auth.User, error) {
	return nil, nil
}
func (r *fakeSettingRepo) UpdateAvatarURL(context.Context, string, string) (*auth.User, error) {
	return nil, nil
}
func (r *fakeSettingRepo) UpdateUserSettings(context.Context, string, map[string]any) (*auth.User, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.updated, nil
}
func (r *fakeSettingRepo) UpdatePasswordHash(context.Context, string, string) error { return nil }

func TestSettingServiceUpdateSettingsSuccess(t *testing.T) {
	repo := &fakeSettingRepo{updated: &auth.User{ID: "u-setting"}}
	svc := NewService(repo)

	updated, err := svc.UpdateMySettings(context.Background(), "u-setting", map[string]any{"locale": "vi-VN"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if updated == nil || updated.ID != "u-setting" {
		t.Fatalf("unexpected updated user: %#v", updated)
	}
}

func TestSettingServiceUpdateSettingsNotFound(t *testing.T) {
	repo := &fakeSettingRepo{updated: nil}
	svc := NewService(repo)

	_, err := svc.UpdateMySettings(context.Background(), "missing", map[string]any{"locale": "vi-VN"})
	if apperrors.KindOf(err) != apperrors.KindNotFound {
		t.Fatalf("expected not_found kind, got %v", apperrors.KindOf(err))
	}
}

func TestSettingServiceUpdateSettingsInternal(t *testing.T) {
	repo := &fakeSettingRepo{err: errors.New("db down")}
	svc := NewService(repo)

	_, err := svc.UpdateMySettings(context.Background(), "u-setting", map[string]any{"locale": "vi-VN"})
	if apperrors.KindOf(err) != apperrors.KindInternal {
		t.Fatalf("expected internal kind, got %v", apperrors.KindOf(err))
	}
}
