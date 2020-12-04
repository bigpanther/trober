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

// Carrier is used by pop to map your carriers database table to your go code.
type Carrier struct {
	ID        uuid.UUID    `json:"id" db:"id"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	CreatedBy uuid.UUID    `json:"created_by" db:"created_by"`
	Name      nulls.String `json:"name" db:"name"`
	Type      string       `json:"type" db:"type"`
	Eta       nulls.Time   `json:"eta" db:"eta"`
	TenantID  uuid.UUID    `json:"tenant_id" db:"tenant_id"`
	Tenant    Tenant       `belongs_to:"tenant"`
}

// String is not required by pop and may be deleted
func (c Carrier) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}

// Carriers is not required by pop and may be deleted
type Carriers []Carrier

// String is not required by pop and may be deleted
func (c Carriers) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (c *Carrier) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: c.Name.String, Name: "Name"},
		&validators.StringIsPresent{Field: c.Type, Name: "Type"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (c *Carrier) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (c *Carrier) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
