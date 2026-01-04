package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/IBM/sarama"
)

const (
	apiURL       = "http://localhost:8080"
	kafkaBrokers = "localhost:9099" // Use external port for localhost
	// Topics from builder.go
	dispatchTopic = "dispatch-events"
	shipmentTopic = "shipment-events"
)

type CreateOrderRequest struct {
	ServiceID     int32               `json:"service_id"`
	ServiceType   string              `json:"service_type"`
	PaymentMethod string              `json:"payment_method"`
	Points        []OrderPointRequest `json:"points"`
	CustomerID    string              `json:"customer_id"`
	DriverID      string              `json:"driver_id,omitempty"` // Optional
}

type OrderPointRequest struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Type    string  `json:"type"`
	Address string  `json:"address"`
	Phone   string  `json:"phone"`
}

type OrderResponse struct {
	ID string `json:"id"`
}

type DispatchEvent struct {
	OrderID        string `json:"order_id"`
	DispatchStatus string `json:"dispatch_status"`
}

type ShipmentEvent struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

func main() {
	// 1. Initialize Kafka Producer
	producer, err := newKafkaProducer()
	if err != nil {
		log.Fatalf("❌ Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()
	fmt.Println("✅ Kafka Producer initialized")

	// 2. Create Order via API
	fmt.Println("🚀 Creating Order...")
	orderID, err := createOrder()
	if err != nil {
		log.Fatalf("❌ Failed to create order: %v", err)
	}
	fmt.Printf("✅ Order Created! ID: %s\n", orderID)

	// 3. Dispatch Order via Kafka
	fmt.Println("\n⏳ Waiting 2 seconds before Dispatching...")
	time.Sleep(2 * time.Second)

	dispatchEvent := DispatchEvent{
		OrderID:        orderID,
		DispatchStatus: "success",
	}
	if err := publishMessage(producer, dispatchTopic, orderID, dispatchEvent); err != nil {
		log.Fatalf("❌ Failed to publish dispatch event: %v", err)
	}
	fmt.Println("✅ Dispatched event published to Kafka!")

	// 4. Deliver Order via Kafka
	fmt.Println("\n⏳ Waiting 5 seconds before Delivering (simulating transit)...")
	time.Sleep(5 * time.Second)

	deliveryEvent := ShipmentEvent{
		OrderID: orderID,
		Status:  "DELIVERED",
	}
	if err := publishMessage(producer, shipmentTopic, orderID, deliveryEvent); err != nil {
		log.Fatalf("❌ Failed to publish shipment event: %v", err)
	}
	fmt.Println("✅ Delivered event published to Kafka!")

	fmt.Println("\n🎉 Full Flow Test Completed Successfully!")
}

func newKafkaProducer() (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	return sarama.NewSyncProducer([]string{kafkaBrokers}, config)
}

func publishMessage(producer sarama.SyncProducer, topic, key string, message interface{}) error {
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(jsonBytes),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return err
	}

	fmt.Printf("📤 Message sent to topic: %s, partition: %d, offset: %d\n", topic, partition, offset)
	return nil
}

func createOrder() (string, error) {
	req := CreateOrderRequest{
		ServiceID:     1,
		ServiceType:   "transport",
		PaymentMethod: "cash",
		CustomerID:    "cust_123",
		Points: []OrderPointRequest{
			{Lat: 10.7769, Lng: 106.7009, Type: "pickup", Address: "SGN", Phone: "0909000000"},
			{Lat: 10.8231, Lng: 106.6297, Type: "dropoff", Address: "Home", Phone: "0909000000"},
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	reqObj, err := http.NewRequest("POST", apiURL+"/orders", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	reqObj.Header.Set("Content-Type", "application/json")
	reqObj.Header.Set("x-user-id", "cust_123")
	reqObj.Header.Set("x-user-role", "customer")

	client := &http.Client{}
	resp, err := client.Do(reqObj)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status: %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data OrderResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err == nil && result.Data.ID != "" {
		return result.Data.ID, nil
	}

	var order OrderResponse
	if err := json.Unmarshal(body, &order); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %s", string(body))
	}
	if order.ID != "" {
		return order.ID, nil
	}

	return "", fmt.Errorf("could not parse order ID from: %s", string(body))
}
