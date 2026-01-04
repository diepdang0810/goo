package entity

type ServiceVO struct {
	ID                 int32    `json:"id"`
	Type               string   `json:"type"`
	Name               string   `json:"name,omitempty"`
	PricingMode        string   `json:"pricing_mode,omitempty"`
	SchedulingMinMax   int      `json:"scheduling_min_max,omitempty"`
	AutoCancelInterval int      `json:"auto_cancel_interval,omitempty"`
	DriverLockTime     int      `json:"driver_lock_time,omitempty"`
	Addons             []string `json:"addons,omitempty"`
	TravelMode         string   `json:"travel_mode,omitempty"`
	Enable             bool     `json:"enable,omitempty"`
}
