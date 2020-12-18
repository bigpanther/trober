package models
// UserRole represents the UserRole enum
type UserRole string
const(
	// UserRoleSuperAdmin represents SuperAdmin UserRole
	UserRoleSuperAdmin UserRole = "SuperAdmin"
	// UserRoleAdmin represents Admin UserRole
	UserRoleAdmin UserRole = "Admin"
	// UserRoleBackOffice represents BackOffice UserRole
	UserRoleBackOffice UserRole = "BackOffice"
	// UserRoleDriver represents Driver UserRole
	UserRoleDriver UserRole = "Driver"
	// UserRoleCustomer represents Customer UserRole
	UserRoleCustomer UserRole = "Customer"
	// UserRoleNone represents None UserRole
	UserRoleNone UserRole = "None"
)
var allowedUserRole [6]UserRole = [6]UserRole{
	UserRoleSuperAdmin,
	UserRoleAdmin,
	UserRoleBackOffice,
	UserRoleDriver,
	UserRoleCustomer,
	UserRoleNone,
}
// IsValidUserRole validates if the input is a UserRole
func IsValidUserRole(s string) bool{
	t := UserRole(s)
	return UserRoleSuperAdmin == t || UserRoleAdmin == t || UserRoleBackOffice == t || UserRoleDriver == t || UserRoleCustomer == t || UserRoleNone == t
}
