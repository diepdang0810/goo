package user

import (
	userHandler "go1/internal/api/handlers/user"
	"go1/internal/modules/user/infrastructure/caching"
	"go1/internal/modules/user/infrastructure/message_queue"
	"go1/internal/modules/user/infrastructure/repository"
	"go1/internal/modules/user/usecase"
	"go1/pkg/kafka"
	"go1/pkg/redis"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Init(router *gin.Engine, db *pgxpool.Pool, redisClient *redis.RedisClient, kafkaProducer *kafka.KafkaProducer) {
	repo := repository.NewPostgresUserRepository(db)
	cache := caching.NewRedisUserCache(redisClient)
	event := message_queue.NewKafkaUserEvent(kafkaProducer)

	uc := usecase.NewUserUsecase(repo, cache, event)
	handler := userHandler.NewUserHandler(uc)
	
	userHandler.RegisterRoutes(router, handler)
}
