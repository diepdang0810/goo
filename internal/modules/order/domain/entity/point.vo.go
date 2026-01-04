package entity

// Point Value Object
type PointVO struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Address string  `json:"address,omitempty"`
	Type    string  `json:"type,omitempty"` // "pickup", "dropoff", "stop"
	Order   int     `json:"order,omitempty"`
}
