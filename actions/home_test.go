package actions

import (
	"net/http"
)

func (as *ActionSuite) Test_HomeHandler() {
	res := as.JSON("/").Get()
	as.Equal(http.StatusOK, res.Code)
	var home = &home{}
	res.Bind(home)
	as.Equal(home.Message, "Welcome to Trober!")
	as.Equal(home.Version, "dev")
	as.Equal(home.Commit, "dev")
}

func (as *ActionSuite) Test_AppInfoHandler() {
	res := as.JSON("/appinfo").Get()
	as.Equal(http.StatusOK, res.Code)
	var appInfo = &appInfo{}
	res.Bind(appInfo)
	as.Equal(appInfo.MinVersion, "0.0.1")
}
