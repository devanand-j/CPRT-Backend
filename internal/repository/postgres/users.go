package postgres

import (
	"context"
	"errors"

	"cprt-lis/internal/domain"

	"github.com/jackc/pgx/v5"

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
		SELECT
			u.id::text,
			u.user_uuid::text,
			u.group_id,
			COALESCE(u.login_id, u.username),
			COALESCE(NULLIF(u.user_name, ''), u.username),
			u.username,
			u.password_hash,
			COALESCE(ag.group_code, ''),
			COALESCE(ag.group_name, ''),
			u.status,
			u.created_at,
			u.last_login
		FROM users u
		LEFT JOIN account_groups ag ON ag.id = u.group_id
		WHERE u.login_id = $1 OR u.username = $1 OR u.email = $1
		LIMIT 1
	`
	var user domain.User
	row := r.pool.QueryRow(ctx, query, username)
	if err := row.Scan(
		&user.ID,
		&user.UserUUID,
		&user.GroupID,
		&user.LoginID,
		&user.DisplayName,
		&user.Username,
		&user.PasswordHash,
		&user.GroupCode,
		&user.GroupName,
		&user.Status,
		&user.CreatedAt,
		&user.LastLogin,
	); err != nil {
		return domain.User{}, errors.New("user not found")
	}
	user.Role = user.GroupCode
	return user, nil
}

func (r *UserRepository) GetAll(ctx context.Context) ([]domain.User, error) {
	const query = `
		SELECT
			u.id::text,
			u.user_uuid::text,
			u.group_id,
			COALESCE(u.login_id, u.username),
			COALESCE(NULLIF(u.user_name, ''), u.username),
			u.username,
			COALESCE(ag.group_code, ''),
			COALESCE(ag.group_name, ''),
			u.status,
			u.created_at,
			u.last_login
		FROM users u
		LEFT JOIN account_groups ag ON ag.id = u.group_id
		ORDER BY u.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []domain.User{}
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID,
			&user.UserUUID,
			&user.GroupID,
			&user.LoginID,
			&user.DisplayName,
			&user.Username,
			&user.GroupCode,
			&user.GroupName,
			&user.Status,
			&user.CreatedAt,
			&user.LastLogin,
		); err != nil {
			return nil, err
		}
		user.Role = user.GroupCode
		users = append(users, user)
	}
	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, id string, accountGroupID *int64, status, passwordHash, updatedBy *string) error {
	if accountGroupID == nil && status == nil && passwordHash == nil && updatedBy == nil {
		return errors.New("no fields to update")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const updateUserQuery = `
		UPDATE users u
		SET
			status = COALESCE($2, u.status),
			password_hash = COALESCE($3, u.password_hash),
			updated_by = COALESCE($4, u.updated_by),
			group_id = COALESCE($5, u.group_id),
			updated_at = NOW()
		WHERE u.id = $1::bigint
	`

	result, err := tx.Exec(ctx, updateUserQuery, id, status, passwordHash, updatedBy, accountGroupID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return tx.Commit(ctx)
}
