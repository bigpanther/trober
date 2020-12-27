package actions

import (
	"net/http"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
)

func selfGet(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.JSON(loggedInUser(c)))
}

func selfGetTenant(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	tenant := &models.Tenant{}

	if err := tx.Find(tenant, loggedInUser(c).TenantID); err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	return c.Render(http.StatusOK, r.JSON(tenant))
}
