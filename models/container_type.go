package models
// ContainerType represents the ContainerType enum
type ContainerType string
const(
	// ContainerTypeIncoming represents Incoming ContainerType
	ContainerTypeIncoming ContainerType = "Incoming"
	// ContainerTypeOutGoing represents OutGoing ContainerType
	ContainerTypeOutGoing ContainerType = "OutGoing"
)
var allowedContainerType [2]ContainerType = [2]ContainerType{
	ContainerTypeIncoming,
	ContainerTypeOutGoing,
}
// IsValidContainerType validates if the input is a ContainerType
func IsValidContainerType(s string) bool{
	t := ContainerType(s)
	return ContainerTypeIncoming == t || ContainerTypeOutGoing == t
}