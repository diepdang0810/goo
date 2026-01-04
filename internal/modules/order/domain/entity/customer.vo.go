package entity

// CustomerVO Value Object
type CustomerVO struct {
	ID    string `json:"id"`
	Phone string `json:"phone,omitempty"`
	Name  string `json:"name,omitempty"`
}
