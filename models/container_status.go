package models
// ContainerStatus represents the ContainerStatus enum
type ContainerStatus string
const (
	// ContainerStatusUnassigned represents Unassigned ContainerStatus
	ContainerStatusUnassigned ContainerStatus = "Unassigned"
	// ContainerStatusInTransit represents InTransit ContainerStatus
	ContainerStatusInTransit ContainerStatus = "InTransit"
	// ContainerStatusArrived represents Arrived ContainerStatus
	ContainerStatusArrived ContainerStatus = "Arrived"
	// ContainerStatusAssigned represents Assigned ContainerStatus
	ContainerStatusAssigned ContainerStatus = "Assigned"
	// ContainerStatusAccepted represents Accepted ContainerStatus
	ContainerStatusAccepted ContainerStatus = "Accepted"
	// ContainerStatusRejected represents Rejected ContainerStatus
	ContainerStatusRejected ContainerStatus = "Rejected"
	// ContainerStatusLoaded represents Loaded ContainerStatus
	ContainerStatusLoaded ContainerStatus = "Loaded"
	// ContainerStatusUnloaded represents Unloaded ContainerStatus
	ContainerStatusUnloaded ContainerStatus = "Unloaded"
	// ContainerStatusAbandoned represents Abandoned ContainerStatus
	ContainerStatusAbandoned ContainerStatus = "Abandoned"
)

var allowedContainerStatus [9]ContainerStatus = [9]ContainerStatus{
	ContainerStatusUnassigned,
	ContainerStatusInTransit,
	ContainerStatusArrived,
	ContainerStatusAssigned,
	ContainerStatusAccepted,
	ContainerStatusRejected,
	ContainerStatusLoaded,
	ContainerStatusUnloaded,
	ContainerStatusAbandoned,
}

// IsValidContainerStatus validates if the input is a ContainerStatus
func IsValidContainerStatus(s string) bool {
	t := ContainerStatus(s)
	return ContainerStatusUnassigned == t || ContainerStatusInTransit == t || ContainerStatusArrived == t || ContainerStatusAssigned == t || ContainerStatusAccepted == t || ContainerStatusRejected == t || ContainerStatusLoaded == t || ContainerStatusUnloaded == t || ContainerStatusAbandoned == t
}
