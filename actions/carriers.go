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
// Model: Singular (Carrier)
// DB Table: Plural (carriers)
// Resource: Plural (Carriers)
// Path: Plural (/carriers)
// View Template Folder: Plural (/templates/carriers/)

// carriersList gets all Carriers. This function is mapped to the path
// GET /carriers
func carriersList(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}
	carrierName := strings.Trim(c.Param("name"), " '")
	carrierType := c.Param("type")
	carriers := &models.Carriers{}
	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())

	if carrierName != "" {
		if len(carrierName) < 2 {
			return c.Render(http.StatusOK, r.JSON(carriers))
		}
		q = q.Where("name ILIKE ?", fmt.Sprintf("%s%%", carrierName))
	}
	if carrierType != "" {
		q = q.Where("type = ?", carrierType)
	}
	// Retrieve all Carriers from the DB
	// Order by the ones that are going to arrive soon
	if err := q.Scope(restrictedScope(c)).Order(fmt.Sprintf("GREATEST(-(now()-eta),(now()-eta)), %s", orderByCreatedAtDesc)).All(carriers); err != nil {
		return err
	}

	return c.Render(200, r.JSON(carriers))

}

// carriersShow gets the data for one Carrier. This function is mapped to
// the path GET /carriers/{carrier_id}
func carriersShow(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)
	if loggedInUser.IsNotActive() {
		return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(http.StatusText(http.StatusNotFound), fmt.Sprint(http.StatusNotFound), errNotFound)))
	}
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

// carriersCreate adds a Carrier to the DB. This function is mapped to the
// path POST /carriers
func carriersCreate(c buffalo.Context) error {

	var loggedInUser = loggedInUser(c)
	if !loggedInUser.IsAtLeastBackOffice() {
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

// carriersUpdate changes a Carrier in the DB. This function is mapped to
// the path PUT /carriers/{carrier_id}
func carriersUpdate(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)
	if !loggedInUser.IsAtLeastBackOffice() {
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

// carriersDestroy deletes a Carrier from the DB. This function is mapped
// to the path DELETE /carriers/{carrier_id}
func carriersDestroy(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)
	if !loggedInUser.IsAtLeastBackOffice() {
		return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(http.StatusText(http.StatusNotFound), fmt.Sprint(http.StatusNotFound), errNotFound)))
	}
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
