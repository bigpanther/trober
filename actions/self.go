package actions

import (
	"net/http"

	"github.com/bigpanther/trober/firebase"
	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
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

func selfPostDeviceRegister(f firebase.Firebase) func(c buffalo.Context) error {
	return func(c buffalo.Context) error {
		deviceB := deviceID{}
		if err := c.Bind(&deviceB); err != nil {
			c.Logger().Errorf("error binding deviceid: %v\n", err)

			return err
		}
		if err := f.SubscribeToTopics(c, loggedInUser(c), deviceB.Token); err != nil {
			return err
		}
		c.Response().WriteHeader(http.StatusNoContent)
		return nil
	}
}

func selfPostDeviceRemove(f firebase.Firebase) func(c buffalo.Context) error {
	return func(c buffalo.Context) error {
		deviceB := deviceID{}
		if err := c.Bind(&deviceB); err != nil {
			c.Logger().Errorf("error binding deviceid: %v\n", err)

			return err
		}
		if err := f.UnSubscribeToTopics(c, loggedInUser(c), deviceB.Token); err != nil {
			return err
		}
		c.Response().WriteHeader(http.StatusNoContent)
		return nil
	}
}

type deviceID struct {
	Token string `json:"token"`
}
