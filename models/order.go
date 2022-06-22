package models

import (
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// Order is used by pop to map your orders database table to your go code.
type Order struct {
	ID               uuid.UUID    `json:"id" db:"id"`
	CreatedAt        time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at" db:"updated_at"`
	CreatedBy        uuid.UUID    `json:"created_by" db:"created_by"`
	TenantID         uuid.UUID    `json:"tenant_id" db:"tenant_id"`
	CustomerID       uuid.UUID    `json:"customer_id" db:"customer_id"`
	CarrierID        nulls.UUID   `json:"carrier_id" db:"carrier_id"`
	TerminalID       nulls.UUID   `json:"terminal_id" db:"terminal_id"`
	SerialNumber     string       `json:"serial_number" db:"serial_number"`
	Status           string       `json:"status" db:"status"`
	Tenant           *Tenant      `belongs_to:"tenant" json:"-"`
	Customer         *Customer    `belongs_to:"customer" json:"customer,omitempty"`
	Carrier          *Carrier     `has_one:"carrier" json:"carrier,omitempty"`
	Terminal         *Terminal    `has_one:"terminal" json:"terminal,omitempty"`
	Shipments        Shipments    `has_many:"shipments" json:"shipments,omitempty"`
	Eta              nulls.Time   `json:"eta" db:"eta"`
	SoNumber         nulls.String `json:"so_number" db:"so_number"`
	Shipline         nulls.String `json:"shipline" db:"shipline"`
	PickupCharges    nulls.Int    `json:"pickup_charges" db:"pickup_charges"`
	PickupCost       nulls.Int    `json:"pickup_cost" db:"pickup_cost"`
	DropoffCharges   nulls.Int    `json:"dropoff_charges" db:"dropoff_charges"`
	DropoffCost      nulls.Int    `json:"dropoff_cost" db:"dropoff_cost"`
	Rld              nulls.String `json:"rld" db:"rld"`
	Erd              nulls.Time   `json:"erd" db:"erd"`
	Docco            nulls.Time   `json:"docco" db:"docco"`
	Lfd              nulls.Time   `json:"lfd" db:"lfd"`
	ContainterStatus nulls.String `json:"container_status" db:"container_status"`
	ShipmentCount    int          `json:"shipmentCount" db:"-"`
	Type             string       `json:"type" db:"type"`
}

// Orders is not required by pop and may be deleted
type Orders []Order

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (o *Order) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: o.SerialNumber, Name: "SerialNumber"},
		&validators.FuncValidator{Fn: func() bool {
			return IsValidOrderStatus(o.Status)
		}, Field: o.Status, Name: "Status"},
		&validators.FuncValidator{Fn: func() bool {
			return IsValidShipmentType(o.Type)
		}, Field: o.Type, Name: "Type"},
	), nil
}
