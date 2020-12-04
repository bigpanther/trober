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

// Tenant is used by pop to map your tenants database table to your go code.
type Tenant struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy nulls.UUID `json:"created_by" db:"created_by"`
	Name      string     `json:"name" db:"name"`
	Type      string     `json:"type" db:"type"`
	Code      string     `json:"code" db:"code"`
}

// String is not required by pop and may be deleted
func (t Tenant) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Tenants is not required by pop and may be deleted
type Tenants []Tenant

// String is not required by pop and may be deleted
func (t Tenants) String() string {
	jt, _ := json.Marshal(t)
	return string(jt)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (t *Tenant) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: t.Name, Name: "Name"},
		&validators.StringIsPresent{Field: t.Type, Name: "Type"},
		&validators.StringIsPresent{Field: t.Code, Name: "Code"},
		&validators.FuncValidator{Fn: func() bool {
			return t.Type == tenantTypeSystem || t.Type == tenantTypeTest || t.Type == tenantTypeProduction
		}, Field: t.Type, Name: "Type"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (t *Tenant) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (t *Tenant) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

const (
	tenantTypeSystem     = "System"
	tenantTypeTest       = "Test"
	tenantTypeProduction = "Production"
)
