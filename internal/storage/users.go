package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type User struct {
	ID          string    `json:"id"`
	Email       *string   `json:"email,omitempty"`
	Phone       *string   `json:"phone,omitempty"`
	DisplayName *string   `json:"display_name,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserWithPassword struct {
	User
	PasswordHash string
}

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")

func EnsureUsersSchema(ctx context.Context, db *Postgres) error {
	if db == nil {
		return errors.New("DATABASE_URL not set")
	}
	pool, err := db.Pool(ctx)
	if err != nil {
		return err
	}

	// Minimal schema based on goen-docs User schema + password hash.
	// We generate IDs in Go; no extensions required.
	_, err = pool.Exec(ctx, `
create table if not exists users (
  id text primary key,
  email text null,
  phone text null,
  display_name text null,
  status text not null default 'active',
  password_hash text not null,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  constraint users_email_or_phone_chk check (email is not null or phone is not null)
);
create unique index if not exists users_email_uq on users (lower(email)) where email is not null;
create unique index if not exists users_phone_uq on users (phone) where phone is not null;
`)
	return err
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func CreateUser(ctx context.Context, db *Postgres, u User, passwordHash string) (*User, error) {
	pool, err := db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	var email *string
	if u.Email != nil {
		n := normalizeEmail(*u.Email)
		email = &n
	}

	row := pool.QueryRow(ctx, `
insert into users (id, email, phone, display_name, status, password_hash, created_at, updated_at)
values ($1, $2, $3, $4, $5, $6, $7, $8)
returning id, email, phone, display_name, status, created_at, updated_at
`, u.ID, email, u.Phone, u.DisplayName, u.Status, passwordHash, u.CreatedAt, u.UpdatedAt)

	var out User
	var outEmail *string
	if err := row.Scan(&out.ID, &outEmail, &out.Phone, &out.DisplayName, &out.Status, &out.CreatedAt, &out.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}
	if outEmail != nil {
		n := normalizeEmail(*outEmail)
		out.Email = &n
	}
	return &out, nil
}

func GetUserByID(ctx context.Context, db *Postgres, id string) (*User, error) {
	pool, err := db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `select id, email, phone, display_name, status, created_at, updated_at from users where id=$1`, id)
	var out User
	var outEmail *string
	if err := row.Scan(&out.ID, &outEmail, &out.Phone, &out.DisplayName, &out.Status, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if outEmail != nil {
		n := normalizeEmail(*outEmail)
		out.Email = &n
	}
	return &out, nil
}

func GetUserWithPasswordByLogin(ctx context.Context, db *Postgres, login string) (*UserWithPassword, error) {
	pool, err := db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	login = strings.TrimSpace(login)
	if login == "" {
		return nil, ErrUserNotFound
	}

	// Simple heuristic: if it looks like an email, match by lower(email), else match by phone.
	isEmail := strings.Contains(login, "@")
	var row pgx.Row
	if isEmail {
		n := normalizeEmail(login)
		row = pool.QueryRow(ctx, `select id, email, phone, display_name, status, password_hash, created_at, updated_at from users where lower(email)=lower($1)`, n)
	} else {
		row = pool.QueryRow(ctx, `select id, email, phone, display_name, status, password_hash, created_at, updated_at from users where phone=$1`, login)
	}

	var out UserWithPassword
	var outEmail *string
	if err := row.Scan(&out.ID, &outEmail, &out.Phone, &out.DisplayName, &out.Status, &out.PasswordHash, &out.CreatedAt, &out.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if outEmail != nil {
		n := normalizeEmail(*outEmail)
		out.Email = &n
	}
	return &out, nil
}
