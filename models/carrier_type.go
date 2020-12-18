package models
// CarrierType represents the CarrierType enum
type CarrierType string
const(
	// CarrierTypeAir represents Air CarrierType
	CarrierTypeAir CarrierType = "Air"
	// CarrierTypeRail represents Rail CarrierType
	CarrierTypeRail CarrierType = "Rail"
	// CarrierTypeVessel represents Vessel CarrierType
	CarrierTypeVessel CarrierType = "Vessel"
	// CarrierTypeRoad represents Road CarrierType
	CarrierTypeRoad CarrierType = "Road"
)
var allowedCarrierType [4]CarrierType = [4]CarrierType{
	CarrierTypeAir,
	CarrierTypeRail,
	CarrierTypeVessel,
	CarrierTypeRoad,
}
// IsValidCarrierType validates if the input is a CarrierType
func IsValidCarrierType(s string) bool{
	t := CarrierType(s)
	return CarrierTypeAir == t || CarrierTypeRail == t || CarrierTypeVessel == t || CarrierTypeRoad == t
}
