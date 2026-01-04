package order

import (
	orderHandler "go1/internal/api/handlers/order"
	"go1/internal/shared/order/application"
	"go1/internal/shared/order/application/validator"
	"go1/internal/shared/order/infrastructure/gateway"
	"go1/internal/shared/order/infrastructure/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Init(router *gin.Engine, db *pgxpool.Pool) {
	// Infrastructure
	repo := repository.NewPostgresOrderRepository(db)

	pricingGw := gateway.NewPricingGateway()
	serviceGw := gateway.NewServiceGateway()
	paymentGw := gateway.NewPaymentGateway()
	locationGw := gateway.NewLocationGateway()

	// Application
	rideValidator := validator.NewRideOrderValidator()

	mapper := application.NewOrderMapper()
	service := application.NewOrderService(
		repo,
		mapper,
		pricingGw,
		serviceGw,
		paymentGw,
		locationGw,
		rideValidator,
	)

	// Presentation
	handler := orderHandler.NewOrderHandler(service)
	orderHandler.RegisterRoutes(router, handler)
}
