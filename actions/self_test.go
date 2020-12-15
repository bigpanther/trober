package actions

import (
	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/httptest"
)

func (as *ActionSuite) Test_SefGet() {
	user := as.getLoggedInUser("nicoleta")
	req := as.setupRequest(user, "/self")
	res := req.Get()
	as.Equal(200, res.Code)
	var self = &models.User{}
	res.Bind(self)
	as.Equal("nicoleta", self.Username)
}

func (as *ActionSuite) Test_SefGetTenant() {
	user := as.getLoggedInUser("nicoleta")
	req := as.setupRequest(user, "/self/tenant")
	res := req.Get()
	as.Equal(200, res.Code)
	var tenant = &models.Tenant{}
	res.Bind(tenant)
	as.Equal("Big Panther Technologies Inc.", tenant.Name)
}

func (as *ActionSuite) getLoggedInUser(username string) *models.User {
	as.LoadFixture("Tenant bootstrap")
	var user = &models.User{}
	err := as.DB.Where("username=?", username).First(user)
	as.NoError(err)
	as.NotZero(user.ID)
	return user
}

func (as *ActionSuite) setupRequest(user *models.User, route string) *httptest.JSON {
	req := as.JSON(route)
	req.Headers[xToken] = user.Username
	return req
}
