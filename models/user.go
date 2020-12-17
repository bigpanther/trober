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

// User is used by pop to map your users database table to your go code.
type User struct {
	ID         uuid.UUID    `json:"id" db:"id"`
	CreatedAt  time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at" db:"updated_at"`
	CreatedBy  nulls.UUID   `json:"created_by" db:"created_by"`
	Name       string       `json:"name" db:"name"`
	Username   string       `json:"username" db:"username"`
	Email      string       `json:"email" db:"email"`
	DeviceID   nulls.String `json:"device_id" db:"device_id"`
	Role       string       `json:"role" db:"role"`
	TenantID   uuid.UUID    `json:"tenant_id" db:"tenant_id"`
	CustomerID nulls.UUID   `json:"customer_id" db:"customer_id"`
	Tenant     *Tenant      `belongs_to:"tenant" json:"-"`
	Customer   *Customer    `belongs_to:"customer" json:"customer,omitempty"`
}

// String is not required by pop and may be deleted
func (u User) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Users is not required by pop and may be deleted
type Users []User

// String is not required by pop and may be deleted
func (u Users) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (u *User) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: u.Name, Name: "Name"},
		&validators.StringIsPresent{Field: u.Username, Name: "Username"},
		&validators.StringIsPresent{Field: u.Role, Name: "Role"},
		&validators.EmailIsPresent{Name: "Email", Field: u.Email, Message: "Email format not valid"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *User) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// IsSuperAdmin checks if a user can work across tenants
func (u *User) IsSuperAdmin() bool {
	return u.Role == userRoleSuperAdmin
}

// IsAdmin checks if a user can work across tenants
func (u *User) IsAdmin() bool {
	return u.Role == userRoleAdmin
}

// IsBackOffice checks if a user can work across tenants
func (u *User) IsBackOffice() bool {
	return u.Role == userRoleBackOffice
}

// IsDriver checks if a user is a driver
func (u *User) IsDriver() bool {
	return u.Role == userRoleDriver
}

// IsCustomer checks if a user is a customer
func (u *User) IsCustomer() bool {
	return u.Role == userRoleCustomer
}

// IsNotActive Mostly for newly created users who have not been assigned a tenant
func (u *User) IsNotActive() bool {
	return u.Role == userRoleNone || u.TenantID == uuid.Nil
}

// AtleastBackOffice checks if a user has at least Back Office access
func (u *User) AtleastBackOffice() bool {
	return u.Role == userRoleSuperAdmin || u.Role == userRoleAdmin || u.Role == userRoleBackOffice
}

// AtleastTenantBackOffice checks if a user has at least Back Office access
func (u *User) AtleastTenantBackOffice() bool {
	return u.Role == userRoleAdmin || u.Role == userRoleBackOffice
}

const (
	userRoleSuperAdmin = "SuperAdmin"
	userRoleAdmin      = "Admin"
	userRoleBackOffice = "BackOffice"
	userRoleCustomer   = "Customer"
	userRoleDriver     = "Driver"
	userRoleNone       = "None"
)
