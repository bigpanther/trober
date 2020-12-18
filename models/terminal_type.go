package models
// TerminalType represents the TerminalType enum
type TerminalType string
const (
	// TerminalTypeRail represents Rail TerminalType
	TerminalTypeRail TerminalType = "Rail"
	// TerminalTypePort represents Port TerminalType
	TerminalTypePort TerminalType = "Port"
	// TerminalTypeWarehouse represents Warehouse TerminalType
	TerminalTypeWarehouse TerminalType = "Warehouse"
	// TerminalTypeYard represents Yard TerminalType
	TerminalTypeYard TerminalType = "Yard"
	// TerminalTypeCustom represents Custom TerminalType
	TerminalTypeCustom TerminalType = "Custom"
)

var allowedTerminalType [5]TerminalType = [5]TerminalType{
	TerminalTypeRail,
	TerminalTypePort,
	TerminalTypeWarehouse,
	TerminalTypeYard,
	TerminalTypeCustom,
}

// IsValidTerminalType validates if the input is a TerminalType
func IsValidTerminalType(s string) bool {
	t := TerminalType(s)
	return TerminalTypeRail == t || TerminalTypePort == t || TerminalTypeWarehouse == t || TerminalTypeYard == t || TerminalTypeCustom == t
}
