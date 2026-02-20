package postgres

import (
	"context"
	"errors"

	"cprt-lis/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (domain.User, error) {
	const query = `
		SELECT id, user_uuid, username, email, phone, password_hash, status, created_at
		FROM users
		WHERE username = $1
	`
	var user domain.User
	row := r.pool.QueryRow(ctx, query, username)
	if err := row.Scan(&user.ID, &user.UserUUID, &user.Username, &user.Email, &user.Phone, &user.PasswordHash, &user.Status, &user.CreatedAt); err != nil {
		return domain.User{}, errors.New("user not found")
	}
	return user, nil
}
