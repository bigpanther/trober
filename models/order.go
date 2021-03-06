package models

import (
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// Order is used by pop to map your orders database table to your go code.
type Order struct {
	ID           uuid.UUID `json:"id" db:"id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy    uuid.UUID `json:"created_by" db:"created_by"`
	TenantID     uuid.UUID `json:"tenant_id" db:"tenant_id"`
	CustomerID   uuid.UUID `json:"customer_id" db:"customer_id"`
	SerialNumber string    `json:"serial_number" db:"serial_number"`
	Status       string    `json:"status" db:"status"`
	Tenant       *Tenant   `belongs_to:"tenant" json:"-"`
	Customer     *Customer `belongs_to:"customer" json:"customer,omitempty"`
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
	), nil
}
