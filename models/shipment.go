package models

import (
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// Shipment is used by pop to map your shipments database table to your go code.
type Shipment struct {
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

// Shipments is not required by pop and may be deleted
type Shipments []Shipment

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (c *Shipment) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: c.SerialNumber, Name: "SerialNumber"},
		&validators.FuncValidator{Fn: func() bool {
			return IsValidShipmentStatus(c.Status)
		}, Field: c.Status, Name: "Status"},
		&validators.FuncValidator{Fn: func() bool {
			// Value can be null
			return !c.Size.Valid || IsValidShipmentSize(c.Size.String)
		}, Field: c.Size.String, Name: "Size"},
		&validators.FuncValidator{Fn: func() bool {
			return IsValidShipmentType(c.Type)
		}, Field: c.Type, Name: "Type"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (c *Shipment) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (c *Shipment) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// IsAssigned checks if a shipment is assigned to a driver
func (c *Shipment) IsAssigned() bool {
	return ShipmentStatus(c.Status) == ShipmentStatusAssigned && c.DriverID.UUID != uuid.Nil
}

// IsRejected checks if a shipment is assigned to a driver
func (c *Shipment) IsRejected() bool {
	return ShipmentStatus(c.Status) == ShipmentStatusRejected
}
