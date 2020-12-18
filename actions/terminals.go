package actions

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
)

// This file is generated by Buffalo. It offers a basic structure for
// adding, editing and deleting a page. If your model is more
// complex or you need more than the basic implementation you need to
// edit this file.

// Following naming logic is implemented in Buffalo:
// Model: Singular (Terminal)
// DB Table: Plural (terminals)
// Resource: Plural (Terminals)
// Path: Plural (/terminals)
// View Template Folder: Plural (/templates/terminals/)

// terminalsList gets all Terminals. This function is mapped to the path
// GET /terminals
func terminalsList(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)
	if loggedInUser.IsNotActive() {
		return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(notFound, fmt.Sprint(http.StatusNotFound), errNotFound)))
	}
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}
	terminalName := strings.Trim(c.Param("name"), " '")
	terminalType := c.Param("type")
	terminals := &models.Terminals{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())

	if terminalName != "" {
		if len(terminalName) < 2 {
			return c.Render(http.StatusOK, r.JSON(terminals))
		}
		q = q.Where("name ILIKE ?", fmt.Sprintf("%s%%", terminalName))
	}
	if terminalType != "" {
		q = q.Where("type = ?", terminalType)
	}
	// Retrieve all Terminals from the DB
	if err := q.Scope(restrictedScope(c)).Order(orderByCreatedAtDesc).All(terminals); err != nil {
		return err
	}

	return c.Render(200, r.JSON(terminals))

}

// terminalsShow gets the data for one Terminal. This function is mapped to
// the path GET /terminals/{terminal_id}
func terminalsShow(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Terminal
	terminal := &models.Terminal{}

	// To find the Terminal the parameter terminal_id is used.
	if err := tx.Scope(restrictedScope(c)).Find(terminal, c.Param("terminal_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	return c.Render(200, r.JSON(terminal))

}

// terminalsCreate adds a Terminal to the DB. This function is mapped to the
// path POST /terminals
func terminalsCreate(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)
	if !loggedInUser.IsAtleastBackOffice() {
		return models.ErrNotFound
	}
	// Allocate an empty Terminal
	terminal := &models.Terminal{}

	// Bind terminal to the html form elements
	if err := c.Bind(terminal); err != nil {
		return err
	}

	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	if !loggedInUser.IsSuperAdmin() || terminal.TenantID == uuid.Nil {
		terminal.TenantID = loggedInUser.TenantID
	}
	terminal.CreatedBy = loggedInUser.ID
	terminal.CreatedAt = time.Now().UTC()
	terminal.UpdatedAt = time.Now().UTC()
	// Validate the data from the html form
	verrs, err := tx.ValidateAndCreate(terminal)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}

	return c.Render(http.StatusCreated, r.JSON(terminal))

}

// terminalsUpdate changes a Terminal in the DB. This function is mapped to
// the path PUT /terminals/{terminal_id}
func terminalsUpdate(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)
	if !loggedInUser.IsAtleastBackOffice() {
		return models.ErrNotFound
	}
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Terminal
	terminal := &models.Terminal{}

	if err := tx.Scope(restrictedScope(c)).Find(terminal, c.Param("terminal_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	// Bind Terminal to the html form elements
	if err := c.Bind(terminal); err != nil {
		return err
	}
	terminal.UpdatedAt = time.Now().UTC()
	verrs, err := tx.ValidateAndUpdate(terminal)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}
	return c.Render(http.StatusOK, r.JSON(terminal))
}

// terminalsDestroy deletes a Terminal from the DB. This function is mapped
// to the path DELETE /terminals/{terminal_id}
func terminalsDestroy(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)
	if !loggedInUser.IsAtleastBackOffice() {
		return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(notFound, fmt.Sprint(http.StatusNotFound), errNotFound)))
	}
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Terminal
	terminal := &models.Terminal{}

	// To find the Terminal the parameter terminal_id is used.
	if err := tx.Scope(restrictedScope(c)).Find(terminal, c.Param("terminal_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(terminal); err != nil {
		return err
	}

	return c.Render(http.StatusOK, r.JSON(terminal))

}
