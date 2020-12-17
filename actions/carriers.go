package actions

import (
	"net/http"
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
// Model: Singular (Carrier)
// DB Table: Plural (carriers)
// Resource: Plural (Carriers)
// Path: Plural (/carriers)
// View Template Folder: Plural (/templates/carriers/)

// CarriersResource is the resource for the Carrier model
type CarriersResource struct {
	buffalo.Resource
}

// List gets all Carriers. This function is mapped to the path
// GET /carriers
func (v CarriersResource) List(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	carriers := &models.Carriers{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())

	// Retrieve all Carriers from the DB
	if err := q.Scope(restrictedScope(c)).Order(orderByCreatedAtDesc).All(carriers); err != nil {
		return err
	}

	return c.Render(200, r.JSON(carriers))

}

// Show gets the data for one Carrier. This function is mapped to
// the path GET /carriers/{carrier_id}
func (v CarriersResource) Show(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Carrier
	carrier := &models.Carrier{}

	// To find the Carrier the parameter carrier_id is used.
	if err := tx.Scope(restrictedScope(c)).Find(carrier, c.Param("carrier_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	return c.Render(200, r.JSON(carrier))

}

// Create adds a Carrier to the DB. This function is mapped to the
// path POST /carriers
func (v CarriersResource) Create(c buffalo.Context) error {

	var loggedInUser = loggedInUser(c)
	if !loggedInUser.IsAtleastBackOffice() {
		return models.ErrNotFound
	}
	// Allocate an empty Carrier
	carrier := &models.Carrier{}

	// Bind carrier to the html form elements
	if err := c.Bind(carrier); err != nil {
		return err
	}

	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	if !loggedInUser.IsSuperAdmin() || carrier.TenantID == uuid.Nil {
		carrier.TenantID = loggedInUser.TenantID
	}
	carrier.CreatedBy = loggedInUser.ID
	carrier.CreatedAt = time.Now().UTC()
	carrier.UpdatedAt = time.Now().UTC()
	// Validate the data from the html form
	verrs, err := tx.ValidateAndCreate(carrier)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}

	return c.Render(http.StatusCreated, r.JSON(carrier))

}

// Update changes a Carrier in the DB. This function is mapped to
// the path PUT /carriers/{carrier_id}
func (v CarriersResource) Update(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)
	if !loggedInUser.IsAtleastBackOffice() {
		return models.ErrNotFound
	}
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Carrier
	carrier := &models.Carrier{}

	if err := tx.Scope(restrictedScope(c)).Find(carrier, c.Param("carrier_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	// Bind Carrier to the html form elements
	if err := c.Bind(carrier); err != nil {
		return err
	}
	carrier.UpdatedAt = time.Now().UTC()

	verrs, err := tx.ValidateAndUpdate(carrier)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}

	return c.Render(http.StatusOK, r.JSON(carrier))

}

// Destroy deletes a Carrier from the DB. This function is mapped
// to the path DELETE /carriers/{carrier_id}
func (v CarriersResource) Destroy(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Carrier
	carrier := &models.Carrier{}

	// To find the Carrier the parameter carrier_id is used.
	if err := tx.Scope(restrictedScope(c)).Find(carrier, c.Param("carrier_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(carrier); err != nil {
		return err
	}

	return c.Render(http.StatusOK, r.JSON(carrier))

}
