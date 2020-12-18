package models
// ContainerSize represents the ContainerSize enum
type ContainerSize string
const(
	// ContainerSize40ST represents 40ST ContainerSize
	ContainerSize40ST ContainerSize = "40ST"
	// ContainerSize20ST represents 20ST ContainerSize
	ContainerSize20ST ContainerSize = "20ST"
	// ContainerSize40HC represents 40HC ContainerSize
	ContainerSize40HC ContainerSize = "40HC"
	// ContainerSize40HW represents 40HW ContainerSize
	ContainerSize40HW ContainerSize = "40HW"
	// ContainerSizeCustom represents Custom ContainerSize
	ContainerSizeCustom ContainerSize = "Custom"
)
var allowedContainerSize [5]ContainerSize = [5]ContainerSize{
	ContainerSize40ST,
	ContainerSize20ST,
	ContainerSize40HC,
	ContainerSize40HW,
	ContainerSizeCustom,
}
// IsValidContainerSize validates if the input is a ContainerSize
func IsValidContainerSize(s string) bool{
	t := ContainerSize(s)
	return ContainerSize40ST == t || ContainerSize20ST == t || ContainerSize40HC == t || ContainerSize40HW == t || ContainerSizeCustom == t
}
