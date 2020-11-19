package actions

import (
	"net/http"
	"os"

	"github.com/gobuffalo/buffalo"
)

var (
	version = os.Getenv("TROBER_VERSION")
	commit  = os.Getenv("TROBER_COMMIT")
)

func HomeHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.JSON(map[string]string{
		"message": "Welcome to Trober!",
		"version": version,
		"commit":  commit,
	}))
}
