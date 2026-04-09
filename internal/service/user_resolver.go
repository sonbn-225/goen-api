package service

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
)

// ResolveUserByLoginTx identifies a user by their login string (email, phone, or username).
// It returns a UserWithPassword entity if found, supporting transactional context.
func ResolveUserByLoginTx(ctx context.Context, tx pgx.Tx, repo interfaces.UserRepository, login string) (*entity.UserWithPassword, error) {
	login = strings.TrimSpace(login)
	if login == "" {
		return nil, apperr.BadRequest("missing_login", "login credential is required")
	}

	if strings.Contains(login, "@") {
		return repo.FindUserByEmailTx(ctx, tx, strings.ToLower(login))
	} else if len(login) > 0 && login[0] != '+' && !utils.IsNumeric(login) {
		return repo.FindUserByUsernameTx(ctx, tx, login)
	} else {
		return repo.FindUserByPhoneTx(ctx, tx, login)
	}
}
