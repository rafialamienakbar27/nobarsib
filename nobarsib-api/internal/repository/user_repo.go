package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nobarsib/nobarsib-api/internal/domain"
)

type UserRepo struct{ db *pgxpool.Pool }

func NewUserRepo(db *pgxpool.Pool) *UserRepo { return &UserRepo{db: db} }

const userColumns = `id, COALESCE(email,''), COALESCE(phone,''),
    COALESCE(password_hash,''), COALESCE(full_name,''), role, is_active,
    last_login_at, created_at`

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Email, &u.Phone, &u.PasswordHash, &u.FullName,
		&u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	const q = `
INSERT INTO app_user (email, phone, password_hash, full_name, role, is_active)
VALUES (NULLIF($1,''), NULLIF($2,''), NULLIF($3,''), NULLIF($4,''), $5, $6)
RETURNING id, created_at`

	err := r.db.QueryRow(ctx, q, strings.ToLower(u.Email), u.Phone, u.PasswordHash,
		u.FullName, u.Role, u.IsActive).Scan(&u.ID, &u.CreatedAt)
	if isUniqueViolation(err) {
		return fmt.Errorf("%w: email atau nomor telepon sudah terdaftar", domain.ErrConflict)
	}
	if isCheckViolation(err) {
		return fmt.Errorf("%w: minimal email atau nomor telepon harus diisi", domain.ErrInvalidInput)
	}
	if err != nil {
		return fmt.Errorf("simpan user: %w", err)
	}
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM app_user WHERE id = $1`, id))
}

// GetByEmail dipakai saat login. Email disimpan apa adanya tapi dicocokkan
// tanpa memperhatikan huruf besar-kecil.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return scanUser(r.db.QueryRow(ctx,
		`SELECT `+userColumns+` FROM app_user WHERE lower(email) = lower($1)`, email))
}

func (r *UserRepo) TouchLogin(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE app_user SET last_login_at = now() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("catat waktu login: %w", err)
	}
	return nil
}
