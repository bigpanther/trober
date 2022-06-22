package actions

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
)

// Following naming logic is implemented in Buffalo:
// Model: Singular (Customer)
// DB Table: Plural (customers)
// Resource: Plural (Customers)
// Path: Plural (/customers)

// customersList gets all Customers. This function is mapped to the path
// GET /customers
func customersList(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	customerName := c.Param("name")

	customers := &models.Customers{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())
	if customerName != "" {
		if len(customerName) < 2 {
			return c.Render(http.StatusOK, r.JSON(customers))
		}
		q = q.Where("name ILIKE ?", fmt.Sprintf("%%%s%%", customerName))
	}
	// Retrieve all Customers from the DB
	if err := q.Scope(restrictedScope(c)).Order(orderByCreatedAtDesc).All(customers); err != nil {
		return err
	}

	return c.Render(http.StatusOK, r.JSON(customers))

}

// customersShow gets the data for one Customer. This function is mapped to
// the path GET /customers/{customer_id}
func customersShow(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)
	customerID := c.Param("customer_id")
	if loggedInUser.IsCustomer() && (!loggedInUser.CustomerID.Valid || loggedInUser.CustomerID.UUID.String() != customerID) {
		return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(http.StatusText(http.StatusNotFound), fmt.Sprint(http.StatusNotFound), errNotFound)))

	}

	tx := c.Value("tx").(*pop.Connection)

	customer := &models.Customer{}

	if err := tx.Scope(restrictedScope(c)).Find(customer, customerID); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	return c.Render(http.StatusOK, r.JSON(customer))

}

// customersCreate adds a Customer to the DB. This function is mapped to the
// path POST /customers
func customersCreate(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)

	customer := &models.Customer{}

	// Bind customer to request body
	if err := c.Bind(customer); err != nil {
		c.Logger().Errorf("error binding customer: %v\n", err)
		return err
	}

	tx := c.Value("tx").(*pop.Connection)

	customer.TenantID = loggedInUser.TenantID
	customer.CreatedBy = nulls.NewUUID(loggedInUser.ID)

	verrs, err := tx.ValidateAndCreate(customer)
	if err != nil {
		c.Logger().Errorf("error creating customer: %v\n", err)
		return err
	}

	if verrs.HasAny() {
		c.Logger().Errorf("customer create errors: %v\n", verrs.String())
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}

	return c.Render(http.StatusCreated, r.JSON(customer))

}

// customersUpdate changes a Customer in the DB. This function is mapped to
// the path PUT /customers/{customer_id}
func customersUpdate(c buffalo.Context) error {

	tx := c.Value("tx").(*pop.Connection)

	customer := &models.Customer{}
	if err := tx.Scope(restrictedScope(c)).Find(customer, c.Param("customer_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	newCustomer := &models.Customer{}
	// Bind customer to request body
	if err := c.Bind(newCustomer); err != nil {
		c.Logger().Errorf("error binding customer: %v\n", err)

		return err
	}
	if newCustomer.Name != customer.Name {
		customer.UpdatedAt = time.Now().UTC()
		customer.Name = newCustomer.Name
	} else {
		return c.Render(http.StatusOK, r.JSON(customer))
	}
	verrs, err := tx.ValidateAndUpdate(customer)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}

	return c.Render(http.StatusOK, r.JSON(customer))

}

// customersDestroy deletes a Customer from the DB. This function is mapped
// to the path DELETE /customers/{customer_id}
func customersDestroy(c buffalo.Context) error {

	tx := c.Value("tx").(*pop.Connection)

	customer := &models.Customer{}

	if err := tx.Scope(restrictedScope(c)).Find(customer, c.Param("customer_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(customer); err != nil {
		return err
	}
	c.Response().WriteHeader(http.StatusNoContent)
	return nil
}
