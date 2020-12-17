package actions

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
)

// Following naming logic is implemented in Buffalo:
// Model: Singular (Tenant)
// DB Table: Plural (tenants)
// Resource: Plural (Tenants)
// Path: Plural (/tenants)
// View Template Folder: Plural (/templates/tenants/)
const (
	notFound             = "Not found"
	orderByCreatedAtDesc = "created_at DESC"
)

var errNotFound = errors.New(notFound)

// tenantsList gets all Tenants. This function is mapped to the path
// GET /tenants
func tenantsList(c buffalo.Context) error {
	// Get the DB connection from the context
	if !loggedInUser(c).IsSuperAdmin() {
		return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(notFound, fmt.Sprint(http.StatusNotFound), errNotFound)))
	}
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	tenants := &models.Tenants{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())

	// Retrieve all Tenants from the DB
	if err := q.Scope(restrictedScope(c)).Order(orderByCreatedAtDesc).All(tenants); err != nil {
		return err
	}
	return c.Render(http.StatusOK, r.JSON(tenants))
}

// tenantsShow gets the data for one Tenant. This function is mapped to
// the path GET /tenants/{tenant_id}
func tenantsShow(c buffalo.Context) error {
	tenantID := c.Param("tenant_id")
	if !loggedInUser(c).IsSuperAdmin() && (loggedInUser(c).IsNotActive() || loggedInUser(c).TenantID.String() != tenantID) {
		return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(notFound, fmt.Sprint(http.StatusNotFound), errNotFound)))
	}
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Tenant
	tenant := &models.Tenant{}

	// To find the Tenant the parameter tenant_id is used.
	if err := tx.Find(tenant, tenantID); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	return c.Render(http.StatusOK, r.JSON(tenant))

}

// tenantsCreate adds a Tenant to the DB. This function is mapped to the
// path POST /tenants
func tenantsCreate(c buffalo.Context) error {
	if !loggedInUser(c).IsSuperAdmin() {
		return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(notFound, fmt.Sprint(http.StatusNotFound), errNotFound)))
	}
	// Allocate an empty Tenant
	tenant := &models.Tenant{}

	// Bind tenant to the html form elements
	if err := c.Bind(tenant); err != nil {
		return err
	}
	tenant.CreatedBy = nulls.NewUUID(loggedInUser(c).ID)
	tenant.CreatedAt = time.Now().UTC()
	tenant.UpdatedAt = time.Now().UTC()

	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Validate the data from the html form
	verrs, err := tx.ValidateAndCreate(tenant)
	if err != nil {
		return err
	}

	if verrs.HasAny() {

		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))

	}

	return c.Render(http.StatusCreated, r.JSON(tenant))

}

// tenantsUpdate changes a Tenant in the DB. This function is mapped to
// the path PUT /tenants/{tenant_id}
func tenantsUpdate(c buffalo.Context) error {
	if !loggedInUser(c).IsSuperAdmin() {
		return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(notFound, fmt.Sprint(http.StatusNotFound), errNotFound)))
	}
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Tenant
	tenant := &models.Tenant{}

	if err := tx.Scope(restrictedScope(c)).Find(tenant, c.Param("tenant_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	// Bind Tenant to the html form elements
	if err := c.Bind(tenant); err != nil {
		return err
	}
	tenant.UpdatedAt = time.Now().UTC()
	verrs, err := tx.ValidateAndUpdate(tenant)
	if err != nil {
		return err
	}

	if verrs.HasAny() {

		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))

	}

	return c.Render(http.StatusOK, r.JSON(tenant))

}

// tenantsDestroy deletes a Tenant from the DB. This function is mapped
// to the path DELETE /tenants/{tenant_id}
func tenantsDestroy(c buffalo.Context) error {
	if !loggedInUser(c).IsSuperAdmin() {
		return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(notFound, fmt.Sprint(http.StatusNotFound), errNotFound)))
	}
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Tenant
	tenant := &models.Tenant{}

	// To find the Tenant the parameter tenant_id is used.
	if err := tx.Find(tenant, c.Param("tenant_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(tenant); err != nil {
		return err
	}
	return c.Render(http.StatusOK, r.JSON(tenant))
}
