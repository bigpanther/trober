package models

import (
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// Carrier is used by pop to map your carriers database table to your go code.
type Carrier struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy uuid.UUID  `json:"created_by" db:"created_by"`
	Name      string     `json:"name" db:"name"`
	Type      string     `json:"type" db:"type"`
	Eta       nulls.Time `json:"eta" db:"eta"`
	TenantID  uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Tenant    *Tenant    `belongs_to:"tenant" json:"-"`
}

// Carriers is not required by pop and may be deleted
type Carriers []Carrier

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (c *Carrier) Validate(tx *pop.Connection) (*validate.Errors, error) {
	c.truncateEta()
	return validate.Validate(
		&validators.StringIsPresent{Field: c.Name, Name: "Name"},
		&validators.FuncValidator{Fn: func() bool {
			return IsValidCarrierType(c.Type)
		}, Field: c.Type, Name: "Type"},
	), nil
}

// truncateEta converts eta to with a minute precision
func (c *Carrier) truncateEta() {
	if c.Eta.Valid {
		c.Eta = nulls.NewTime(c.Eta.Time.UTC().Truncate(time.Minute))
	}
}
