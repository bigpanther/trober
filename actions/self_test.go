package actions

import (
	"testing"

	"github.com/bigpanther/trober/models"
)

func (as *ActionSuite) Test_SelfGetTenant() {
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

func (as *ActionSuite) Test_SelfGet() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username string
		f        func(u *models.User) bool
	}{
		{"mane", func(u *models.User) bool {
			return u.IsAtleastTenantBackOffice() && u.IsAtleastBackOffice() && u.IsBackOffice()
		}},
		{"firmino", func(u *models.User) bool {
			return u.IsAtleastTenantBackOffice() && u.IsAtleastBackOffice() && u.IsAdmin()
		}},
		{"allan", func(u *models.User) bool {
			return u.IsNotActive()
		}},
		{"salah", func(u *models.User) bool {
			return u.IsDriver()
		}},
		{"klopp", func(u *models.User) bool {
			return u.IsAtleastBackOffice() && u.IsSuperAdmin()
		}},
		{"adidas", func(u *models.User) bool {
			return u.IsCustomer()
		}},
	}

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, "/self")
			res := req.Get()
			as.Equal(200, res.Code)
			var self = &models.User{}
			res.Bind(self)
			as.True(test.f(self))
		})
	}
}
