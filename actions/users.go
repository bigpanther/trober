package actions

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
)

// Following naming logic is implemented in Buffalo:
// Model: Singular (User)
// DB Table: Plural (users)
// Resource: Plural (Users)
// Path: Plural (/users)

// usersList gets all Users. This function is mapped to the path
// GET /users
func usersList(c buffalo.Context) error {

	tx := c.Value("tx").(*pop.Connection)
	userName := strings.Trim(c.Param("name"), " '")
	userRole := c.Param("role")
	users := &models.Users{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())

	if userName != "" {
		if len(userName) < 2 {
			return c.Render(http.StatusOK, r.JSON(users))
		}
		q = q.Where("name ILIKE ?", fmt.Sprintf("%s%%", userName))
	}
	if userRole != "" {
		q = q.Where("role = ?", userRole)
	}

	// Retrieve all Users from the DB
	if err := q.Scope(restrictedScope(c)).Order(orderByCreatedAtDesc).All(users); err != nil {
		return err
	}
	return c.Render(http.StatusOK, r.JSON(users))
}

// usersShow gets the data for one User. This function is mapped to
// the path GET /users/{user_id}
func usersShow(c buffalo.Context) error {

	tx := c.Value("tx").(*pop.Connection)

	user := &models.User{}
	var populatedFields = []string{"Customer"}

	if err := tx.Eager(populatedFields...).Scope(restrictedScope(c)).Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	return c.Render(http.StatusOK, r.JSON(user))
}

// usersCreate adds a User to the DB. This function is mapped to the
// path POST /users
func usersCreate(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)

	user := &models.User{}
	// Bind user to request body
	if err := c.Bind(user); err != nil {
		return err
	}
	if user.IsSuperAdmin() {
		return c.Render(http.StatusBadRequest, r.JSON(models.NewCustomError(http.StatusText(http.StatusBadRequest), fmt.Sprint(http.StatusBadRequest), errors.New("User Role value is not valid"))))
	}

	tx := c.Value("tx").(*pop.Connection)
	if err := checkCustomerUser(c, tx, user); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}
	if !loggedInUser.IsSuperAdmin() || user.TenantID == uuid.Nil {
		user.TenantID = loggedInUser.TenantID
	}
	user.Username = fmt.Sprintf("invited-%d", rand.Int())
	user.CreatedBy = nulls.NewUUID(loggedInUser.ID)
	verrs, err := tx.ValidateAndCreate(user)
	if err != nil {
		return err
	}
	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}
	return c.Render(http.StatusCreated, r.JSON(user))
}

// usersUpdate changes a User in the DB. This function is mapped to
// the path PUT /users/{user_id}
func usersUpdate(c buffalo.Context) error {

	tx := c.Value("tx").(*pop.Connection)

	user := &models.User{}
	if err := tx.Scope(restrictedScope(c)).Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	if user.IsSuperAdmin() {
		return c.Render(http.StatusBadRequest, r.JSON(models.NewCustomError(http.StatusText(http.StatusBadRequest), fmt.Sprint(http.StatusBadRequest), errors.New("updating superuser not allowed"))))
	}
	newUser := &models.User{}
	// Bind user to request body
	if err := c.Bind(newUser); err != nil {
		return err
	}
	var loggedInUser = loggedInUser(c)
	if user.ID == loggedInUser.ID {
		// Cannot change self role
		newUser.Role = user.Role
	}

	if newUser.Name != user.Name || newUser.Role != user.Role {
		user.UpdatedAt = time.Now().UTC()
		user.Name = newUser.Name
		user.Role = newUser.Role
		if err := checkCustomerUser(c, tx, newUser); err != nil {
			return c.Error(http.StatusBadRequest, err)
		}
		if err := checkEscalation(loggedInUser, newUser); err != nil {
			return c.Render(http.StatusForbidden, r.JSON(models.NewCustomError(err.Error(), http.StatusText(http.StatusForbidden), err)))
		}
	} else {
		return c.Render(http.StatusOK, r.JSON(user))
	}

	verrs, err := tx.ValidateAndUpdate(user)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}

	return c.Render(http.StatusOK, r.JSON(user))

}

// usersDestroy deletes a User from the DB. This function is mapped
// to the path DELETE /users/{user_id}
func usersDestroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	user := &models.User{}

	if err := tx.Scope(restrictedScope(c)).Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(user); err != nil {
		return err
	}
	c.Response().WriteHeader(http.StatusNoContent)
	return nil

}

func checkCustomerUser(c buffalo.Context, tx *pop.Connection, user *models.User) error {
	if user.IsCustomer() {
		customer := &models.Customer{}
		// User must belong to a customer in the same tenant
		err := tx.Scope(restrictedScope(c)).Where("tenant_id = ?", user.TenantID).Find(customer, user.CustomerID)
		if err != nil || user.TenantID != customer.TenantID {
			return errors.New("invalid customer association")
		}
	} else {
		user.CustomerID = nulls.UUID{}
	}
	return nil
}

var errEscalatePrivileges = errors.New("cannot escalate privileges beyond your own role")

func checkEscalation(self *models.User, user *models.User) error {
	if self.IsBackOffice() && (user.Role == models.UserRoleAdmin.String() || user.Role == models.UserRoleSuperAdmin.String()) {
		return errEscalatePrivileges
	}
	if self.IsAdmin() && user.Role == models.UserRoleSuperAdmin.String() {
		return errEscalatePrivileges
	}
	return nil
}
