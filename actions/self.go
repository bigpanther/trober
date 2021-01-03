package actions

import (
	"net/http"

	"github.com/bigpanther/trober/firebase"
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

func selfDeviceRegister(c buffalo.Context) error {
	deviceB := deviceBody{}
	if err := c.Bind(&deviceB); err != nil {
		return err
	}
	if err := firebase.SubscribeToTopics(c, loggedInUser(c), deviceB.DeviceID); err != nil {
		return err
	}
	c.Response().WriteHeader(http.StatusNoContent)
	return nil
}

func selfDeviceRemove(c buffalo.Context) error {
	deviceB := deviceBody{}
	if err := c.Bind(&deviceB); err != nil {
		return err
	}
	if err := firebase.UnSubscribeToTopics(c, loggedInUser(c), deviceB.DeviceID); err != nil {
		return err
	}
	c.Response().WriteHeader(http.StatusNoContent)
	return nil
}

type deviceBody struct {
	DeviceID string `json:"device_id"`
}
