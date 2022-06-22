package actions

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
)

// Following naming logic is implemented in Buffalo:
// Model: Singular (Terminal)
// DB Table: Plural (terminals)
// Resource: Plural (Terminals)
// Path: Plural (/terminals)

// terminalsList gets all Terminals. This function is mapped to the path
// GET /terminals
func terminalsList(c buffalo.Context) error {

	tx := c.Value("tx").(*pop.Connection)
	terminalName := c.Param("name")
	terminalType := c.Param("type")
	terminals := &models.Terminals{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())

	if terminalName != "" {
		if len(terminalName) < 2 {
			return c.Render(http.StatusOK, r.JSON(terminals))
		}
		q = q.Where("name ILIKE ?", fmt.Sprintf("%%%s%%", terminalName))
	}
	if terminalType != "" {
		q = q.Where("type = ?", terminalType)
	}
	// Retrieve all Terminals from the DB
	if err := q.Scope(restrictedScope(c)).Order(orderByCreatedAtDesc).All(terminals); err != nil {
		return err
	}

	return c.Render(http.StatusOK, r.JSON(terminals))

}

// terminalsShow gets the data for one Terminal. This function is mapped to
// the path GET /terminals/{terminal_id}
func terminalsShow(c buffalo.Context) error {

	tx := c.Value("tx").(*pop.Connection)

	terminal := &models.Terminal{}

	if err := tx.Scope(restrictedScope(c)).Find(terminal, c.Param("terminal_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	return c.Render(http.StatusOK, r.JSON(terminal))
}

// terminalsCreate adds a Terminal to the DB. This function is mapped to the
// path POST /terminals
func terminalsCreate(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)

	terminal := &models.Terminal{}

	// Bind terminal to request body
	if err := c.Bind(terminal); err != nil {
		c.Logger().Errorf("error binding terminal: %v\n", err)

		return err
	}

	tx := c.Value("tx").(*pop.Connection)

	terminal.TenantID = loggedInUser.TenantID
	terminal.CreatedBy = loggedInUser.ID

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

	tx := c.Value("tx").(*pop.Connection)

	terminal := &models.Terminal{}
	if err := tx.Scope(restrictedScope(c)).Find(terminal, c.Param("terminal_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	newTerminal := &models.Terminal{}
	// Bind Terminal to request body
	if err := c.Bind(newTerminal); err != nil {
		c.Logger().Errorf("error binding terminal: %v\n", err)

		return err
	}
	if newTerminal.Name != terminal.Name || newTerminal.Type != terminal.Type {
		terminal.UpdatedAt = time.Now().UTC()
		terminal.Name = newTerminal.Name
		terminal.Type = newTerminal.Type
	} else {
		return c.Render(http.StatusOK, r.JSON(terminal))
	}
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

	tx := c.Value("tx").(*pop.Connection)

	terminal := &models.Terminal{}

	if err := tx.Scope(restrictedScope(c)).Find(terminal, c.Param("terminal_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(terminal); err != nil {
		return err
	}
	c.Response().WriteHeader(http.StatusNoContent)
	return nil

}
