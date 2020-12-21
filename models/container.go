package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// Container is used by pop to map your containers database table to your go code.
type Container struct {
	ID              uuid.UUID    `json:"id" db:"id"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at" db:"updated_at"`
	CreatedBy       uuid.UUID    `json:"created_by" db:"created_by"`
	TenantID        uuid.UUID    `json:"tenant_id" db:"tenant_id"`
	CarrierID       nulls.UUID   `json:"carrier_id" db:"carrier_id"`
	TerminalID      nulls.UUID   `json:"terminal_id" db:"terminal_id"`
	OrderID         nulls.UUID   `json:"order_id" db:"order_id"`
	SerialNumber    string       `json:"serial_number" db:"serial_number"`
	Origin          nulls.String `json:"origin" db:"origin"`
	Destination     nulls.String `json:"destination" db:"destination"`
	Lfd             nulls.Time   `json:"lfd" db:"lfd"`
	ReservationTime nulls.Time   `json:"reservation_time" db:"reservation_time"`
	Size            nulls.String `json:"size" db:"size"`
	Type            string       `json:"type" db:"type"`
	Status          string       `json:"status" db:"status"`
	DriverID        nulls.UUID   `json:"driver_id" db:"driver_id"`
	Tenant          *Tenant      `belongs_to:"tenant" json:"-"`
	Terminal        *Terminal    `belongs_to:"terminal"  json:"terminal,omitempty"`
	Carrier         *Carrier     `belongs_to:"carrier" json:"carrier,omitempty"`
	Order           *Order       `belongs_to:"order" json:"order,omitempty"`
	Driver          *User        `belongs_to:"user" json:"driver,omitempty"`
}

// String is not required by pop and may be deleted
func (c Container) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}

// Containers is not required by pop and may be deleted
type Containers []Container

// String is not required by pop and may be deleted
func (c Containers) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (c *Container) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: c.SerialNumber, Name: "SerialNumber"},
		&validators.FuncValidator{Fn: func() bool {
			return IsValidContainerStatus(c.Status)
		}, Field: c.Status, Name: "Status"},
		&validators.FuncValidator{Fn: func() bool {
			// Value can be null
			return !c.Size.Valid || IsValidContainerSize(c.Size.String)
		}, Field: c.Size.String, Name: "Size"},
		&validators.FuncValidator{Fn: func() bool {
			return IsValidContainerType(c.Type)
		}, Field: c.Type, Name: "Type"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (c *Container) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (c *Container) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// IsAssigned checks if a container is assigned to a driver
func (c *Container) IsAssigned() bool {
	return ContainerStatus(c.Status) == ContainerStatusAssigned && c.DriverID.UUID != uuid.Nil
}

// IsRejected checks if a container is assigned to a driver
func (c *Container) IsRejected() bool {
	return ContainerStatus(c.Status) == ContainerStatusRejected
}
