package kafka_test

import (
	"fmt"
	"net/http"
	"strconv"

	"go1/pkg/kafka"
	"go1/pkg/logger"
	"go1/pkg/response"

	"github.com/gin-gonic/gin"
)

type KafkaTestHandler struct {
	producer *kafka.KafkaProducer
}

func NewKafkaTestHandler(producer *kafka.KafkaProducer) *KafkaTestHandler {
	return &KafkaTestHandler{producer: producer}
}

type TestMessage struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
	Topic   string `json:"topic"`
}

type PublishRequest struct {
	Topic   string `json:"topic" binding:"required"`
	Count   int    `json:"count" binding:"required,min=1,max=100"`
	Message string `json:"message"`
}

type PublishResponse struct {
	Success       int      `json:"success"`
	Failed        int      `json:"failed"`
	PublishedIDs  []int    `json:"published_ids"`
	FailedReasons []string `json:"failed_reasons,omitempty"`
}

// PublishTestMessages godoc
// @Summary Publish test messages to Kafka
// @Description Publish multiple test messages to a specified Kafka topic
// @Tags kafka-test
// @Accept json
// @Produce json
// @Param request body PublishRequest true "Publish request"
// @Success 200 {object} PublishResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/kafka-test/publish [post]
func (h *KafkaTestHandler) PublishTestMessages(c *gin.Context) {
	var req PublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	logger.Log.Info("Publishing test messages",
		logger.Field{Key: "topic", Value: req.Topic},
		logger.Field{Key: "count", Value: req.Count},
		logger.Field{Key: "message", Value: req.Message})

	var publishedIDs []int
	var failedReasons []string
	successCount := 0
	failedCount := 0

	for i := 1; i <= req.Count; i++ {
		msg := TestMessage{
			ID:      i,
			Message: fmt.Sprintf("%s (message #%d)", req.Message, i),
			Topic:   req.Topic,
		}

		// Publish struct directly - auto-marshaled to JSON!
		// Key is optional - pass as variadic param
		key := strconv.Itoa(i)
		if err := h.producer.Publish(req.Topic, msg, key); err != nil {
			failedCount++
			failedReasons = append(failedReasons, fmt.Sprintf("ID %d: publish error: %v", i, err))
		} else {
			successCount++
			publishedIDs = append(publishedIDs, i)
		}
	}

	resp := PublishResponse{
		Success:       successCount,
		Failed:        failedCount,
		PublishedIDs:  publishedIDs,
		FailedReasons: failedReasons,
	}

	logger.Log.Info("Publish completed",
		logger.Field{Key: "success", Value: successCount},
		logger.Field{Key: "failed", Value: failedCount})

	response.Success(c, resp)
}

// GetTopics godoc
// @Summary Get available test topics
// @Description Get list of configured topics for testing
// @Tags kafka-test
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/kafka-test/topics [get]
func (h *KafkaTestHandler) GetTopics(c *gin.Context) {
	topics := map[string]interface{}{
		"available_topics": []map[string]string{
			{
				"name":        "test_success",
				"description": "Topic that processes successfully (no errors)",
				"retry":       "disabled",
			},
			{
				"name":        "test_retry",
				"description": "Topic that fails and retries (with DLQ after max attempts)",
				"retry":       "enabled",
				"maxAttempts": "3",
				"backoff":     "2000ms",
			},
		},
		"note": "Use POST /api/v1/kafka-test/publish to send messages",
	}
	response.Success(c, topics)
}
