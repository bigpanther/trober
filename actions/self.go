package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/pop/v5"
	"github.com/shipanther/trober/models"
)

func selfGet(c buffalo.Context) error {
	app.Worker.Perform(worker.Job{
		Queue:   "default",
		Handler: "testWorker",
		Args: worker.Args{
			"user_id": 123,
		},
	})
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
