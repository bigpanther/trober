package grifts

import (
	"trober/models"

	"github.com/markbates/grift/grift"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Seeds a database")
	grift.Add("seed", func(c *grift.Context) error {
		t := &models.Tenant{
			Name: "system",
			Type: "System",
		}
		err := models.DB.Create(t)
		if err != nil {
			return err
		}
		u := &models.User{
			Name:     "HSM",
			Username: "maan.harry@gmail.com",
			TenantID: t.ID,
			Role:     "SuperAdmin",
		}
		return models.DB.Create(u)
	})

})
