package entity

// PaymentVO Value Object
type PaymentVO struct {
	Method      string                 `json:"method"`
	Config      map[string]interface{} `json:"config,omitempty"`
	PaymentType string                 `json:"payment_type,omitempty"`
}
