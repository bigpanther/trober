package grifts

import (
	"github.com/gobuffalo/nulls"
	"github.com/shipanther/trober/models"

	"github.com/markbates/grift/grift"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Seeds a database")
	grift.Add("seed", func(c *grift.Context) error {
		t := &models.Tenant{
			Name: "system",
			Type: "System",
			Code: nulls.NewString("6mapg"),
		}
		err := models.DB.Create(t)
		if err != nil {
			return err
		}
		u := &models.User{
			Name:     "Big Panther",
			Username: "oaxWWvwxFOM0odE8tJqqdZEYdxG3",
			TenantID: t.ID,
			Role:     "SuperAdmin",
			Email:    "info@bigpanther.ca",
		}
		return models.DB.Create(u)
	})

})
