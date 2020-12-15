package actions

import (
	"net/http"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
)

func selfGet(c buffalo.Context) error {
	return c.Render(200, r.JSON(loggedInUser(c)))
}

func selfGetTenant(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}
	tenant := &models.Tenant{}

	if err := tx.Scope(restrictedScope(c)).Find(tenant, loggedInUser(c).TenantID); err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	return c.Render(200, r.JSON(tenant))
}
