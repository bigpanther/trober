package actions

import (
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
	"github.com/shipanther/trober/models"
)

// This file is generated by Buffalo. It offers a basic structure for
// adding, editing and deleting a page. If your model is more
// complex or you need more than the basic implementation you need to
// edit this file.

// Following naming logic is implemented in Buffalo:
// Model: Singular (Order)
// DB Table: Plural (orders)
// Resource: Plural (Orders)
// Path: Plural (/orders)
// View Template Folder: Plural (/templates/orders/)

// OrdersResource is the resource for the Order model
type OrdersResource struct {
	buffalo.Resource
}

// List gets all Orders. This function is mapped to the path
// GET /orders
func (v OrdersResource) List(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	orders := &models.Orders{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())

	// Retrieve all Orders from the DB
	if err := q.Scope(restrictedScope(c)).All(orders); err != nil {
		return err
	}

	return c.Render(200, r.JSON(orders))

}

// Show gets the data for one Order. This function is mapped to
// the path GET /orders/{order_id}
func (v OrdersResource) Show(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Order
	order := &models.Order{}

	// To find the Order the parameter order_id is used.
	if err := tx.Scope(restrictedScope(c)).Find(order, c.Param("order_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	return c.Render(200, r.JSON(order))

}

// Create adds a Order to the DB. This function is mapped to the
// path POST /orders
func (v OrdersResource) Create(c buffalo.Context) error {
	// Allocate an empty Order
	order := &models.Order{}

	// Bind order to the html form elements
	if err := c.Bind(order); err != nil {
		return err
	}

	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}
	var loggedInUser = loggedInUser(c)
	if !loggedInUser.IsSuperAdmin() || order.TenantID == uuid.Nil {
		order.TenantID = loggedInUser.TenantID
	}
	order.CreatedBy = loggedInUser.ID
	order.CreatedAt = time.Now().UTC()
	order.UpdatedAt = time.Now().UTC()

	// Validate the data from the html form
	verrs, err := tx.ValidateAndCreate(order)
	if err != nil {
		return err
	}

	if verrs.HasAny() {

		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))

	}

	return c.Render(http.StatusCreated, r.JSON(order))

}

// Update changes a Order in the DB. This function is mapped to
// the path PUT /orders/{order_id}
func (v OrdersResource) Update(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Order
	order := &models.Order{}

	if err := tx.Scope(restrictedScope(c)).Find(order, c.Param("order_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	// Bind Order to the html form elements
	if err := c.Bind(order); err != nil {
		return err
	}
	order.UpdatedAt = time.Now().UTC()

	verrs, err := tx.ValidateAndUpdate(order)
	if err != nil {
		return err
	}

	if verrs.HasAny() {

		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))

	}

	return c.Render(http.StatusOK, r.JSON(order))

}

// Destroy deletes a Order from the DB. This function is mapped
// to the path DELETE /orders/{order_id}
func (v OrdersResource) Destroy(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Order
	order := &models.Order{}

	// To find the Order the parameter order_id is used.
	if err := tx.Scope(restrictedScope(c)).Find(order, c.Param("order_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(order); err != nil {
		return err
	}

	return c.Render(http.StatusOK, r.JSON(order))

}
