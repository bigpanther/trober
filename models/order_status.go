package models

// AUTOGENERATED BY: HSM GEN

// OrderStatus represents the OrderStatus enum
type OrderStatus string

const (
	// OrderStatusOpen represents Open OrderStatus
	OrderStatusOpen OrderStatus = "Open"
	// OrderStatusAccepted represents Accepted OrderStatus
	OrderStatusAccepted OrderStatus = "Accepted"
	// OrderStatusCancelled represents Cancelled OrderStatus
	OrderStatusCancelled OrderStatus = "Cancelled"
	// OrderStatusInProgress represents InProgress OrderStatus
	OrderStatusInProgress OrderStatus = "InProgress"
	// OrderStatusDelivered represents Delivered OrderStatus
	OrderStatusDelivered OrderStatus = "Delivered"
	// OrderStatusInvoiced represents Invoiced OrderStatus
	OrderStatusInvoiced OrderStatus = "Invoiced"
	// OrderStatusPaymentReceived represents PaymentReceived OrderStatus
	OrderStatusPaymentReceived OrderStatus = "PaymentReceived"
)

var allowedOrderStatus [7]OrderStatus = [7]OrderStatus{
	OrderStatusOpen,
	OrderStatusAccepted,
	OrderStatusCancelled,
	OrderStatusInProgress,
	OrderStatusDelivered,
	OrderStatusInvoiced,
	OrderStatusPaymentReceived,
}

// String returns the string representation of
func (k OrderStatus) String() string {
	return string(k)
}

// IsValidOrderStatus validates if the input is a OrderStatus
func IsValidOrderStatus(s string) bool {
	t := OrderStatus(s)
	return OrderStatusOpen == t || OrderStatusAccepted == t || OrderStatusCancelled == t || OrderStatusInProgress == t || OrderStatusDelivered == t || OrderStatusInvoiced == t || OrderStatusPaymentReceived == t
}
