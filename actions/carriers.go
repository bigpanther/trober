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
// Model: Singular (Carrier)
// DB Table: Plural (carriers)
// Resource: Plural (Carriers)
// Path: Plural (/carriers)

// carriersList gets all Carriers. This function is mapped to the path
// GET /carriers
func carriersList(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	carrierName := c.Param("name")
	carrierType := c.Param("type")
	carriers := &models.Carriers{}
	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())

	if carrierName != "" {
		if len(carrierName) < 2 {
			return c.Render(http.StatusOK, r.JSON(carriers))
		}
		q = q.Where("name ILIKE ?", fmt.Sprintf("%%%s%%", carrierName))
	}
	if carrierType != "" {
		q = q.Where("type = ?", carrierType)
	}
	// Retrieve all Carriers from the DB
	// Order by the ones that are going to arrive soon
	if err := q.Scope(restrictedScope(c)).Order(fmt.Sprintf("GREATEST(-(now()-eta),(now()-eta)), %s", orderByCreatedAtDesc)).All(carriers); err != nil {
		c.Logger().Errorf("error retrieving carriers: %v\n", err)
		return err
	}
	return c.Render(http.StatusOK, r.JSON(carriers))
}

// carriersShow gets the data for one Carrier. This function is mapped to
// the path GET /carriers/{carrier_id}
func carriersShow(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)
	if loggedInUser.IsNotActive() {
		return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(http.StatusText(http.StatusNotFound), fmt.Sprint(http.StatusNotFound), errNotFound)))
	}

	tx := c.Value("tx").(*pop.Connection)

	carrier := &models.Carrier{}

	if err := tx.Scope(restrictedScope(c)).Find(carrier, c.Param("carrier_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	return c.Render(http.StatusOK, r.JSON(carrier))

}

// carriersCreate adds a Carrier to the DB. This function is mapped to the
// path POST /carriers
func carriersCreate(c buffalo.Context) error {

	var loggedInUser = loggedInUser(c)
	if !loggedInUser.IsAtLeastBackOffice() {
		return models.ErrNotFound
	}

	carrier := &models.Carrier{}

	// Bind carrier to request body
	if err := c.Bind(carrier); err != nil {
		c.Logger().Errorf("error binding carrier: %v\n", err)
		return err
	}

	tx := c.Value("tx").(*pop.Connection)

	carrier.TenantID = loggedInUser.TenantID
	carrier.CreatedBy = loggedInUser.ID

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

	tx := c.Value("tx").(*pop.Connection)

	carrier := &models.Carrier{}
	if err := tx.Scope(restrictedScope(c)).Find(carrier, c.Param("carrier_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	newCarrier := &models.Carrier{}
	// Bind carrier to request body
	if err := c.Bind(newCarrier); err != nil {
		c.Logger().Errorf("error binding carrier: %v\n", err)
		return err
	}
	if newCarrier.Name != carrier.Name || newCarrier.Type != carrier.Type || newCarrier.Eta != carrier.Eta {
		carrier.UpdatedAt = time.Now().UTC()
		carrier.Name = newCarrier.Name
		carrier.Type = newCarrier.Type
		carrier.Eta = newCarrier.Eta
	} else {
		return c.Render(http.StatusOK, r.JSON(carrier))
	}

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

	tx := c.Value("tx").(*pop.Connection)

	carrier := &models.Carrier{}

	if err := tx.Scope(restrictedScope(c)).Find(carrier, c.Param("carrier_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(carrier); err != nil {
		return err
	}
	c.Response().WriteHeader(http.StatusNoContent)
	return nil
}
