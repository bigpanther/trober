package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
)

var (
	version = "dev"
	commit  = "dev"
)

func HomeHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.JSON(map[string]string{
		"message": "Welcome to Trober!",
		"version": version,
		"commit":  commit,
	}))
}
