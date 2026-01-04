package entity

// DriverVO Value Object
type DriverVO struct {
	ID    string `json:"id"`
	Phone string `json:"phone,omitempty"`
	Name  string `json:"name,omitempty"`
}
