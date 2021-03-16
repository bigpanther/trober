package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
)

var (
	version                 = "dev"
	commit                  = "dev"
	minimumSupportedVersion = "0.0.1"
)

type home struct {
	Message string `json:"message"`
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

func homeHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.JSON(home{
		Message: "Welcome to Trober!",
		Version: version,
		Commit:  commit,
	}))
}

type appInfo struct {
	MinVersion string `json:"minVersion"`
}

func appInfoHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.JSON(appInfo{
		MinVersion: minimumSupportedVersion,
	}))
}
