package actions

import (
	"testing"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/httptest"
)

func (as *ActionSuite) Test_SefGet() {
	as.LoadFixture("Tenant bootstrap")
	user := as.getLoggedInUser("firmino")
	req := as.setupRequest(user, "/self")
	res := req.Get()
	as.Equal(200, res.Code)
	var self = &models.User{}
	res.Bind(self)
	as.Equal("firmino", self.Username)
}

func (as *ActionSuite) Test_SefGetTenant() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username   string
		tenantName string
	}{
		{"mane", "Big Panther Liverpool"},
		{"firmino", "Big Panther Liverpool"},
		{"allan", "Big Panther Everton"},
		{"salah", "Big Panther Liverpool"},
		{"klopp", "Big Panther Test"},
		{"adidas", "Big Panther Everton"},
	}

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, "/self/tenant")
			res := req.Get()
			as.Equal(200, res.Code)
			var tenant = &models.Tenant{}
			res.Bind(tenant)
			as.Equal(test.tenantName, tenant.Name)
		})
	}
}

func (as *ActionSuite) getLoggedInUser(username string) *models.User {
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
