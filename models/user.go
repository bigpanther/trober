package models

import (
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

// Users is not required by pop and may be deleted
type Users []User

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (u *User) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: u.Name, Name: "Name"},
		&validators.StringIsPresent{Field: u.Username, Name: "Username"},
		&validators.StringIsPresent{Field: u.Role, Name: "Role"},
		&validators.EmailIsPresent{Name: "Email", Field: u.Email},
		&validators.FuncValidator{Fn: func() bool {
			return IsValidUserRole(u.Role)
		}, Field: u.Role, Name: "Role"},
	), nil
}

// IsSuperAdmin checks if a user can work across tenants
func (u *User) IsSuperAdmin() bool {
	return u.Role == UserRoleSuperAdmin.String()
}

// IsAdmin checks if a user can work across tenants
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin.String()
}

// IsBackOffice checks if a user can work across tenants
func (u *User) IsBackOffice() bool {
	return u.Role == UserRoleBackOffice.String()
}

// IsDriver checks if a user is a driver
func (u *User) IsDriver() bool {
	return u.Role == UserRoleDriver.String()
}

// IsCustomer checks if a user is a customer
func (u *User) IsCustomer() bool {
	return u.Role == UserRoleCustomer.String()
}

// IsNotActive Mostly for newly created users who have not been assigned a tenant
func (u *User) IsNotActive() bool {
	return u.Role == UserRoleNone.String() || u.TenantID == uuid.Nil
}

// IsAtLeastBackOffice checks if a user has at least Back Office access
func (u *User) IsAtLeastBackOffice() bool {
	return u.Role == UserRoleSuperAdmin.String() || u.Role == UserRoleAdmin.String() || u.Role == UserRoleBackOffice.String()
}

// IsAtLeastTenantBackOffice checks if a user has at least Back Office access
func (u *User) IsAtLeastTenantBackOffice() bool {
	return u.Role == UserRoleAdmin.String() || u.Role == UserRoleBackOffice.String()
}
