package actions

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
)

func excludeUpdateColumnsDefault() []string {
	return []string{"id", "created_at", "created_by"}
}

// This file is generated by Buffalo. It offers a basic structure for
// adding, editing and deleting a page. If your model is more
// complex or you need more than the basic implementation you need to
// edit this file.

// Following naming logic is implemented in Buffalo:
// Model: Singular (User)
// DB Table: Plural (users)
// Resource: Plural (Users)
// Path: Plural (/users)
// View Template Folder: Plural (/templates/users/)

// usersList gets all Users. This function is mapped to the path
// GET /users
func usersList(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}
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
	return c.Render(200, r.JSON(users))
}

// usersShow gets the data for one User. This function is mapped to
// the path GET /users/{user_id}
func usersShow(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}
	// Allocate an empty User
	user := &models.User{}
	var populatedFields = []string{"Customer"}
	// To find the User the parameter user_id is used.
	if err := tx.Eager(populatedFields...).Scope(restrictedScope(c)).Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	return c.Render(200, r.JSON(user))
}

// usersCreate adds a User to the DB. This function is mapped to the
// path POST /users
func usersCreate(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)
	// Allocate an empty User
	user := &models.User{}
	// Bind user to the html form elements
	if err := c.Bind(user); err != nil {
		return err
	}
	if user.IsSuperAdmin() {
		return c.Render(http.StatusBadRequest, r.JSON(models.NewCustomError(http.StatusText(http.StatusBadRequest), fmt.Sprint(http.StatusBadRequest), errors.New("User Role value is not valid"))))
	}
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}
	if !loggedInUser.IsSuperAdmin() || user.TenantID == uuid.Nil {
		user.TenantID = loggedInUser.TenantID
	}
	user.CreatedBy = nulls.NewUUID(loggedInUser.ID)
	// Validate the data from the html form
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
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty User
	user := &models.User{}

	if err := tx.Scope(restrictedScope(c)).Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	//var originalRole = user.Role
	// Bind User to the html form elements
	if err := c.Bind(user); err != nil {
		return err
	}
	var loggedInUser = loggedInUser(c)
	user.UpdatedAt = time.Now().UTC()
	var excludedColumns = excludeUpdateColumnsDefault()
	if !loggedInUser.IsSuperAdmin() {
		excludedColumns = append(excludedColumns, "tenant_id")
	}
	if user.ID == loggedInUser.ID {
		// Cannot change self role
		excludedColumns = append(excludedColumns, "role")
	}
	verrs, err := tx.ValidateAndUpdate(user, excludedColumns...)
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
	var loggedInUser = loggedInUser(c)
	if !loggedInUser.IsAtLeastBackOffice() {
		return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(http.StatusText(http.StatusNotFound), fmt.Sprint(http.StatusNotFound), errNotFound)))
	}
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty User
	user := &models.User{}

	// To find the User the parameter user_id is used.
	if err := tx.Scope(restrictedScope(c)).Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(user); err != nil {
		return err
	}

	return c.Render(http.StatusOK, r.JSON(user))

}
