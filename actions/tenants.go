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
)

// Following naming logic is implemented in Buffalo:
// Model: Singular (Tenant)
// DB Table: Plural (tenants)
// Resource: Plural (Tenants)
// Path: Plural (/tenants)

const (
	orderByCreatedAtDesc = "created_at DESC"
)

var errNotFound = errors.New(http.StatusText(http.StatusNotFound))

// tenantsList gets all Tenants. This function is mapped to the path
// GET /tenants
func tenantsList(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	tenantName := strings.Trim(c.Param("name"), " '")
	tenantType := c.Param("type")
	tenants := &models.Tenants{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())
	if tenantName != "" {
		if len(tenantName) < 2 {
			return c.Render(http.StatusOK, r.JSON(tenants))
		}
		q = q.Where("name ILIKE ?", fmt.Sprintf("%s%%", tenantName))
	}
	if tenantType != "" {
		q = q.Where("type = ?", tenantType)
	}
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

	tx := c.Value("tx").(*pop.Connection)

	tenant := &models.Tenant{}

	if err := tx.Find(tenant, tenantID); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	return c.Render(http.StatusOK, r.JSON(tenant))

}

// tenantsCreate adds a Tenant to the DB. This function is mapped to the
// path POST /tenants
func tenantsCreate(c buffalo.Context) error {
	if !loggedInUser(c).IsSuperAdmin() {
		return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(http.StatusText(http.StatusNotFound), fmt.Sprint(http.StatusNotFound), errNotFound)))
	}

	tenant := &models.Tenant{}

	// Bind tenant to request body
	if err := c.Bind(tenant); err != nil {
		return err
	}
	tenant.CreatedBy = nulls.NewUUID(loggedInUser(c).ID)

	tx := c.Value("tx").(*pop.Connection)

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

	tx := c.Value("tx").(*pop.Connection)

	tenant := &models.Tenant{}
	if err := tx.Scope(restrictedScope(c)).Find(tenant, c.Param("tenant_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	newTenant := &models.Tenant{}
	// Bind Tenant to request body
	if err := c.Bind(newTenant); err != nil {
		return err
	}
	if newTenant.Name != tenant.Name || newTenant.Type != tenant.Type || newTenant.Code != tenant.Code {
		tenant.UpdatedAt = time.Now().UTC()
		tenant.Name = newTenant.Name
		tenant.Type = newTenant.Type
		tenant.Code = newTenant.Code
	} else {
		return c.Render(http.StatusOK, r.JSON(tenant))
	}
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

	tx := c.Value("tx").(*pop.Connection)

	tenant := &models.Tenant{}

	if err := tx.Find(tenant, c.Param("tenant_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(tenant); err != nil {
		return err
	}
	return c.Render(http.StatusOK, r.JSON(tenant))
}
