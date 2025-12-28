package order

import (
	orderHandler "go1/internal/api/handlers/order"
	"go1/internal/modules/order/infrastructure/repository"
	"go1/internal/modules/order/usecase"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.temporal.io/sdk/client"
)

func Init(router *gin.Engine, db *pgxpool.Pool, temporalClient client.Client) {
	repo := repository.NewPostgresOrderRepository(db)
	uc := usecase.NewOrderUsecase(repo, temporalClient)
	handler := orderHandler.NewOrderHandler(uc)

	orderHandler.RegisterRoutes(router, handler)
}
