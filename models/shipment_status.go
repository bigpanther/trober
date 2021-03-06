package models

// AUTOGENERATED BY: HSM GEN

// ShipmentStatus represents the ShipmentStatus enum
type ShipmentStatus string

const (
	// ShipmentStatusUnassigned represents Unassigned ShipmentStatus
	ShipmentStatusUnassigned ShipmentStatus = "Unassigned"
	// ShipmentStatusInTransit represents InTransit ShipmentStatus
	ShipmentStatusInTransit ShipmentStatus = "InTransit"
	// ShipmentStatusArrived represents Arrived ShipmentStatus
	ShipmentStatusArrived ShipmentStatus = "Arrived"
	// ShipmentStatusAssigned represents Assigned ShipmentStatus
	ShipmentStatusAssigned ShipmentStatus = "Assigned"
	// ShipmentStatusAccepted represents Accepted ShipmentStatus
	ShipmentStatusAccepted ShipmentStatus = "Accepted"
	// ShipmentStatusRejected represents Rejected ShipmentStatus
	ShipmentStatusRejected ShipmentStatus = "Rejected"
	// ShipmentStatusLoaded represents Loaded ShipmentStatus
	ShipmentStatusLoaded ShipmentStatus = "Loaded"
	// ShipmentStatusDelivered represents Delivered ShipmentStatus
	ShipmentStatusDelivered ShipmentStatus = "Delivered"
)

var allowedShipmentStatus [8]ShipmentStatus = [8]ShipmentStatus{
	ShipmentStatusUnassigned,
	ShipmentStatusInTransit,
	ShipmentStatusArrived,
	ShipmentStatusAssigned,
	ShipmentStatusAccepted,
	ShipmentStatusRejected,
	ShipmentStatusLoaded,
	ShipmentStatusDelivered,
}

// String returns the string representation of
func (k ShipmentStatus) String() string {
	return string(k)
}

// IsValidShipmentStatus validates if the input is a ShipmentStatus
func IsValidShipmentStatus(s string) bool {
	t := ShipmentStatus(s)
	return ShipmentStatusUnassigned == t || ShipmentStatusInTransit == t || ShipmentStatusArrived == t || ShipmentStatusAssigned == t || ShipmentStatusAccepted == t || ShipmentStatusRejected == t || ShipmentStatusLoaded == t || ShipmentStatusDelivered == t
}
