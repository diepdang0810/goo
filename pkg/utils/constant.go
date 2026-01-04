package utils

type UserRole string

const (
	UserRoleCustomer UserRole = "user_app"
	UserRoleAdmin    UserRole = "admin_portal"
	UserRoleDriver   UserRole = "driver_app"
)

type MyHeader string

const (
	XUserIDHeader       MyHeader = "X-User-Id"
	XUserAudienceHeader MyHeader = "X-User-Audience"
	XUserPlatformHeader MyHeader = "X-User-Platform"
	XPartnerIDHeader    MyHeader = "X-Partner-Id"
)

type ServiceType string

const (
	ServiceTypeRideTaxi    ServiceType = "RIDE-TAXI"
	ServiceTypeRideHour    ServiceType = "RIDE-HOUR"
	ServiceTypeRideRoute   ServiceType = "RIDE-ROUTE"
	ServiceTypeRideTrip    ServiceType = "RIDE-TRIP"
	ServiceTypeRideShare   ServiceType = "RIDE-SHARE"
	ServiceTypeRideShuttle ServiceType = "RIDE-SHUTTLE"
)

type PayType string

const (
	PayTypeCash     PayType = "cash"
	PayTypePrePaid  PayType = "prepaid"
	PayTypePayLater PayType = "pay_later"
)
