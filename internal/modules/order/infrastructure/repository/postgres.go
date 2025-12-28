package repository

import (
	"context"
	"fmt"

	"go1/internal/modules/order/domain"
	"go1/internal/modules/order/infrastructure/repository/postgres/mapper"
	"go1/internal/modules/order/infrastructure/repository/postgres/model"

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

type postgresOrderRepository struct {
	db PgxPoolIface
}

var errNotFound = fmt.Errorf("not found")

func NewPostgresOrderRepository(db *pgxpool.Pool) domain.OrderRepository {
	return &postgresOrderRepository{db: db}
}

func (r *postgresOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	query := `INSERT INTO orders (user_id, amount, status, workflow_id, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING id, created_at, updated_at`

	var m model.OrderModel
	err := r.db.QueryRow(ctx, query, order.UserID, order.Amount, order.Status, order.WorkflowID).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return fmt.Errorf("postgresOrderRepository.Create: %w", err)
	}

	order.ID = m.ID
	order.CreatedAt = m.CreatedAt
	order.UpdatedAt = m.UpdatedAt

	return nil
}

func (r *postgresOrderRepository) GetByID(ctx context.Context, id int64) (*domain.Order, error) {
	query := `SELECT id, user_id, amount, status, workflow_id, created_at, updated_at FROM orders WHERE id = $1`

	var m model.OrderModel
	err := r.db.QueryRow(ctx, query, id).Scan(&m.ID, &m.UserID, &m.Amount, &m.Status, &m.WorkflowID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errNotFound
		}
		return nil, fmt.Errorf("postgresOrderRepository.GetByID: %w", err)
	}

	return mapper.ToOrderDomain(&m), nil
}

func (r *postgresOrderRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("postgresOrderRepository.UpdateStatus: %w", err)
	}
	return nil
}
