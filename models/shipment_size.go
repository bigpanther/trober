package models

// AUTOGENERATED BY: HSM GEN

// ShipmentSize represents the ShipmentSize enum
type ShipmentSize string

const (
	// ShipmentSize40ST represents 40ST ShipmentSize
	ShipmentSize40ST ShipmentSize = "40ST"
	// ShipmentSize20ST represents 20ST ShipmentSize
	ShipmentSize20ST ShipmentSize = "20ST"
	// ShipmentSize40HC represents 40HC ShipmentSize
	ShipmentSize40HC ShipmentSize = "40HC"
	// ShipmentSize40HW represents 40HW ShipmentSize
	ShipmentSize40HW ShipmentSize = "40HW"
	// ShipmentSizeCustom represents Custom ShipmentSize
	ShipmentSizeCustom ShipmentSize = "Custom"
)

var allowedShipmentSize [5]ShipmentSize = [5]ShipmentSize{
	ShipmentSize40ST,
	ShipmentSize20ST,
	ShipmentSize40HC,
	ShipmentSize40HW,
	ShipmentSizeCustom,
}

// String returns the string representation of
func (k ShipmentSize) String() string {
	return string(k)
}

// IsValidShipmentSize validates if the input is a ShipmentSize
func IsValidShipmentSize(s string) bool {
	t := ShipmentSize(s)
	return ShipmentSize40ST == t || ShipmentSize20ST == t || ShipmentSize40HC == t || ShipmentSize40HW == t || ShipmentSizeCustom == t
}
