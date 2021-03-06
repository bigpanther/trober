package models

// AUTOGENERATED BY: HSM GEN

// TenantType represents the TenantType enum
type TenantType string

const (
	// TenantTypeSystem represents System TenantType
	TenantTypeSystem TenantType = "System"
	// TenantTypeTest represents Test TenantType
	TenantTypeTest TenantType = "Test"
	// TenantTypeProduction represents Production TenantType
	TenantTypeProduction TenantType = "Production"
)

var allowedTenantType [3]TenantType = [3]TenantType{
	TenantTypeSystem,
	TenantTypeTest,
	TenantTypeProduction,
}

// String returns the string representation of
func (k TenantType) String() string {
	return string(k)
}

// IsValidTenantType validates if the input is a TenantType
func IsValidTenantType(s string) bool {
	t := TenantType(s)
	return TenantTypeSystem == t || TenantTypeTest == t || TenantTypeProduction == t
}
