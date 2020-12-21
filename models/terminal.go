package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// Terminal is used by pop to map your terminals database table to your go code.
type Terminal struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`
	Name      string    `json:"name" db:"name"`
	Type      string    `json:"type" db:"type"`
	TenantID  uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Tenant    *Tenant   `belongs_to:"tenant" json:"-"`
}

// String is not required by pop and may be deleted
func (t Terminal) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Terminals is not required by pop and may be deleted
type Terminals []Terminal

// String is not required by pop and may be deleted
func (t Terminals) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (t *Terminal) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: t.Name, Name: "Name"},
		&validators.FuncValidator{Fn: func() bool {
			return IsValidTerminalType(t.Type)
		}, Field: t.Type, Name: "Type"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (t *Terminal) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (t *Terminal) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
