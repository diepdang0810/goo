package shipment

import (
	"go1/pkg/kafka"
	"go1/pkg/response"
	"go1/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ShipmentHandler struct {
	kafkaProducer *kafka.KafkaProducer
}

func NewShipmentHandler(kafkaProducer *kafka.KafkaProducer) *ShipmentHandler {
	return &ShipmentHandler{kafkaProducer: kafkaProducer}
}

func (h *ShipmentHandler) Accept(c *gin.Context) {
	orderIDStr := c.Param("orderId")
	orderID, err := utils.ParseInt(orderIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid Order ID")
		return
	}

	// Send message to Kafka
	err = h.kafkaProducer.Publish("shipment-events", map[string]interface{}{
		"order_id": orderID,
		"status":   "ACCEPTED",
	}, "order-accepted")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Order accepted"})
}

func (h *ShipmentHandler) Delivery(c *gin.Context) {
	orderIDStr := c.Param("orderId")
	orderID, err := utils.ParseInt(orderIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid Order ID")
		return
	}

	// Send message to Kafka
	err = h.kafkaProducer.Publish("shipment-events", map[string]interface{}{
		"order_id": orderID,
		"status":   "DELIVERED",
	}, "order-delivered")
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Order delivered"})
}
