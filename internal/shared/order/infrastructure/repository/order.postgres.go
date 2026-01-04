package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go1/internal/shared/order/domain"
	"go1/internal/shared/order/domain/entity"
	"go1/internal/shared/order/infrastructure/repository/postgres/mapper"
	"go1/internal/shared/order/infrastructure/repository/postgres/model"
	"go1/pkg/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgxPoolIface defines the interface for pgx pool operations
// PgxPoolIface defines the interface for pgx pool operations
type PgxPoolIface interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	Close()
}

type postgresOrderRepository struct {
	db PgxPoolIface
}

var errNotFound = fmt.Errorf("not found")

func NewPostgresOrderRepository(db *pgxpool.Pool) domain.OrderRepository {
	return &postgresOrderRepository{db: db}
}

func (r *postgresOrderRepository) Create(ctx context.Context, order *entity.RideOrderEntity) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO orders (
		id, created_by, status, payment_method, metadata, workflow_id, service_id, service_type, service_name, created_at, updated_at,
		sub_status, promotion_code, fee_id, has_insurance, order_time, completed_time, cancel_time, platform, is_schedule, now_order, now_order_code
	) 
	VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
		$12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
	) 
	RETURNING id, created_at, updated_at`

	metadataBytes, err := json.Marshal(order.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	var m model.OrderModel
	err = tx.QueryRow(ctx, query,
		order.ID,
		order.CreatedBy,
		order.Status,
		order.Payment.Method,
		metadataBytes,
		order.WorkflowID,
		order.Service.ID,
		order.Service.Type,
		order.Service.Name,
		order.CreatedAt,
		order.UpdatedAt,
		utils.EmptyToNil(order.SubStatus),
		utils.EmptyToNil(order.PromotionCode),
		utils.EmptyToNil(order.FeeID),
		order.HasInsurance,
		order.OrderTime,
		order.CompletedTime,
		order.CancelTime,
		order.Platform,
		order.IsSchedule,
		order.NowOrder,
		utils.EmptyToNil(order.NowOrderCode),
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)

	if err != nil {
		return fmt.Errorf("postgresOrderRepository.Create: %w", err)
	}

	// Insert Order Points
	pointQuery := `INSERT INTO order_points (order_id, lat, lng, address, type, ordering) VALUES ($1, $2, $3, $4, $5, $6)`
	for i, p := range order.Points {
		_, err := tx.Exec(ctx, pointQuery, order.ID, p.Lat, p.Lng, p.Address, p.Type, i)
		if err != nil {
			return fmt.Errorf("failed to insert order point: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// fields are already set in order, just update timestamps if returned
	order.CreatedAt = m.CreatedAt
	order.UpdatedAt = m.UpdatedAt

	return nil
}

func (r *postgresOrderRepository) GetByID(ctx context.Context, id string) (*entity.RideOrderEntity, error) {
	query := `SELECT 
		id, created_by, status, payment_method, metadata, workflow_id, service_id, service_type, service_name, created_at, updated_at,
		sub_status, promotion_code, fee_id, has_insurance, order_time, completed_time, cancel_time, platform, is_schedule, now_order, now_order_code
	FROM orders WHERE id = $1`

	var m model.OrderModel
	var subStatus, promotionCode, feeID, nowOrderCode *string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&m.ID,
		&m.CreatedBy,
		&m.Status,
		&m.PaymentMethod,
		&m.Metadata,
		&m.WorkflowID,
		&m.ServiceID,
		&m.ServiceType,
		&m.ServiceName,
		&m.CreatedAt,
		&m.UpdatedAt,
		&subStatus,
		&promotionCode,
		&feeID,
		&m.HasInsurance,
		&m.OrderTime,
		&m.CompletedTime,
		&m.CancelTime,
		&m.Platform,
		&m.IsSchedule,
		&m.NowOrder,
		&nowOrderCode,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errNotFound
		}
		return nil, fmt.Errorf("postgresOrderRepository.GetByID: %w", err)
	}

	if subStatus != nil {
		m.SubStatus = *subStatus
	}
	if promotionCode != nil {
		m.PromotionCode = *promotionCode
	}
	if feeID != nil {
		m.FeeID = *feeID
	}
	if nowOrderCode != nil {
		m.NowOrderCode = *nowOrderCode
	}

	return mapper.ToOrderDomain(&m), nil
}

func (r *postgresOrderRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("postgresOrderRepository.UpdateStatus: %w", err)
	}
	return nil
}

func (r *postgresOrderRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM orders WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("postgresOrderRepository.Delete: %w", err)
	}
	return nil
}

func (r *postgresOrderRepository) Update(ctx context.Context, order *entity.RideOrderEntity) error {
	query := `UPDATE orders SET status = $1, payment_method = $2, metadata = $3, updated_at = $4 WHERE id = $5`
	metadataBytes, err := json.Marshal(order.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	_, err = r.db.Exec(ctx, query, order.Status, order.Payment.Method, metadataBytes, time.Now(), order.ID)
	if err != nil {
		return fmt.Errorf("postgresOrderRepository.Update: %w", err)
	}
	return nil
}
