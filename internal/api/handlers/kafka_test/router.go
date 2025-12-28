package kafka_test

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, handler *KafkaTestHandler) {
	v1 := router.Group("/api/v1")
	{
		kafkaTest := v1.Group("/kafka-test")
		{
			kafkaTest.POST("/publish", handler.PublishTestMessages)
			kafkaTest.GET("/topics", handler.GetTopics)
		}
	}
}
