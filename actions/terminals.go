package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/x/responder"
	"github.com/shipanther/trober/models"
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

// TerminalsResource is the resource for the Terminal model
type TerminalsResource struct {
	buffalo.Resource
}

// List gets all Terminals. This function is mapped to the path
// GET /terminals
func (v TerminalsResource) List(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	terminals := &models.Terminals{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())

	// Retrieve all Terminals from the DB
	if err := q.All(terminals); err != nil {
		return err
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		// Add the paginator to the context so it can be used in the template.
		c.Set("pagination", q.Paginator)

		c.Set("terminals", terminals)
		return c.Render(http.StatusOK, r.HTML("/terminals/index.plush.html"))
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(200, r.JSON(terminals))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(200, r.XML(terminals))
	}).Respond(c)
}

// Show gets the data for one Terminal. This function is mapped to
// the path GET /terminals/{terminal_id}
func (v TerminalsResource) Show(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Terminal
	terminal := &models.Terminal{}

	// To find the Terminal the parameter terminal_id is used.
	if err := tx.Find(terminal, c.Param("terminal_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		c.Set("terminal", terminal)

		return c.Render(http.StatusOK, r.HTML("/terminals/show.plush.html"))
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(200, r.JSON(terminal))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(200, r.XML(terminal))
	}).Respond(c)
}

// Create adds a Terminal to the DB. This function is mapped to the
// path POST /terminals
func (v TerminalsResource) Create(c buffalo.Context) error {
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

	// Validate the data from the html form
	verrs, err := tx.ValidateAndCreate(terminal)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return responder.Wants("html", func(c buffalo.Context) error {
			// Make the errors available inside the html template
			c.Set("errors", verrs)

			// Render again the new.html template that the user can
			// correct the input.
			c.Set("terminal", terminal)

			return c.Render(http.StatusUnprocessableEntity, r.HTML("/terminals/new.plush.html"))
		}).Wants("json", func(c buffalo.Context) error {
			return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
		}).Wants("xml", func(c buffalo.Context) error {
			return c.Render(http.StatusUnprocessableEntity, r.XML(verrs))
		}).Respond(c)
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		// If there are no errors set a success message
		c.Flash().Add("success", T.Translate(c, "terminal.created.success"))

		// and redirect to the show page
		return c.Redirect(http.StatusSeeOther, "/terminals/%v", terminal.ID)
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(http.StatusCreated, r.JSON(terminal))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(http.StatusCreated, r.XML(terminal))
	}).Respond(c)
}

// Update changes a Terminal in the DB. This function is mapped to
// the path PUT /terminals/{terminal_id}
func (v TerminalsResource) Update(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Terminal
	terminal := &models.Terminal{}

	if err := tx.Find(terminal, c.Param("terminal_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	// Bind Terminal to the html form elements
	if err := c.Bind(terminal); err != nil {
		return err
	}

	verrs, err := tx.ValidateAndUpdate(terminal)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return responder.Wants("html", func(c buffalo.Context) error {
			// Make the errors available inside the html template
			c.Set("errors", verrs)

			// Render again the edit.html template that the user can
			// correct the input.
			c.Set("terminal", terminal)

			return c.Render(http.StatusUnprocessableEntity, r.HTML("/terminals/edit.plush.html"))
		}).Wants("json", func(c buffalo.Context) error {
			return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
		}).Wants("xml", func(c buffalo.Context) error {
			return c.Render(http.StatusUnprocessableEntity, r.XML(verrs))
		}).Respond(c)
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		// If there are no errors set a success message
		c.Flash().Add("success", T.Translate(c, "terminal.updated.success"))

		// and redirect to the show page
		return c.Redirect(http.StatusSeeOther, "/terminals/%v", terminal.ID)
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.JSON(terminal))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.XML(terminal))
	}).Respond(c)
}

// Destroy deletes a Terminal from the DB. This function is mapped
// to the path DELETE /terminals/{terminal_id}
func (v TerminalsResource) Destroy(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Terminal
	terminal := &models.Terminal{}

	// To find the Terminal the parameter terminal_id is used.
	if err := tx.Find(terminal, c.Param("terminal_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(terminal); err != nil {
		return err
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		// If there are no errors set a flash message
		c.Flash().Add("success", T.Translate(c, "terminal.destroyed.success"))

		// Redirect to the index page
		return c.Redirect(http.StatusSeeOther, "/terminals")
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.JSON(terminal))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.XML(terminal))
	}).Respond(c)
}
