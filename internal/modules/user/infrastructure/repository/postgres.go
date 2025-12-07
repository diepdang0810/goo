package repository

import (
	"context"
	"fmt"

	"go1/internal/modules/user/domain"
	"go1/internal/modules/user/infrastructure/repository/postgres/mapper"
	"go1/internal/modules/user/infrastructure/repository/postgres/model"
	"go1/pkg/apperrors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgxPoolIface defines the interface for pgx pool operations
type PgxPoolIface interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Close()
}

type postgresUserRepository struct {
	db PgxPoolIface
}

func NewPostgresUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &postgresUserRepository{db: db}
}

func NewPostgresUserRepositoryWithInterface(db PgxPoolIface) domain.UserRepository {
	return &postgresUserRepository{db: db}
}



func (r *postgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (name, email, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id, created_at, updated_at`
	
	var m model.UserModel
	err := r.db.QueryRow(ctx, query, user.Name, user.Email).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return apperrors.ErrEmailAlreadyExists
		}
		return fmt.Errorf("postgresUserRepository.Create: %w", err)
	}
	
	user.ID = m.ID
	user.CreatedAt = m.CreatedAt
	user.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1`
	var m model.UserModel
	err := r.db.QueryRow(ctx, query, id).Scan(&m.ID, &m.Name, &m.Email, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("postgresUserRepository.GetByID: %w", err)
	}
	return mapper.ToDomain(&m), nil
}

func (r *postgresUserRepository) Fetch(ctx context.Context) ([]domain.User, error) {
	query := `SELECT id, name, email, created_at, updated_at FROM users`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("postgresUserRepository.Fetch: %w", err)
	}
	defer rows.Close()

	var users = make([]domain.User, 0)
	for rows.Next() {
		var m model.UserModel
		if err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("postgresUserRepository.Fetch scan: %w", err)
		}
		users = append(users, *mapper.ToDomain(&m))
	}
	return users, nil
}

func (r *postgresUserRepository) DeleteByID(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("postgresUserRepository.DeleteByID: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return apperrors.ErrUserNotFound
	}
	
	return nil
}
