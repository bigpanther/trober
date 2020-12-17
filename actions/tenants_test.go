package actions

import (
	"fmt"
	"testing"

	"github.com/bigpanther/trober/models"
)

func (as *ActionSuite) Test_TenantsResource_List() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
		tenantCount  int
	}{
		{"mane", 404, 0},
		{"firmino", 404, 0},
		{"allan", 404, 0},
		{"salah", 404, 0},
		{"klopp", 200, 3},
		{"adidas", 404, 0},
	}

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, "/tenants")
			res := req.Get()
			as.Equal(test.responseCode, res.Code)
			if test.responseCode == 200 {
				var tenants = &models.Tenants{}
				res.Bind(tenants)
				as.Equal(len(*tenants), test.tenantCount)
			}
		})
	}

}

func (as *ActionSuite) Test_TenantsResource_Show() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
		tenantName   string
	}{
		{"mane", 200, "Big Panther Liverpool"},
		{"firmino", 200, "Big Panther Liverpool"},
		{"allan", 404, "Big Panther Everton"},
		{"salah", 200, "Big Panther Liverpool"},
		{"klopp", 200, "Big Panther Test"},
		{"adidas", 200, "Big Panther Everton"},
	}

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, fmt.Sprintf("/tenants/%s", user.TenantID))
			res := req.Get()
			as.Equal(test.responseCode, res.Code)
			if test.responseCode == 200 {
				var tenant = &models.Tenant{}
				res.Bind(tenant)
				as.Equal(test.tenantName, tenant.Name)
			}
		})
	}
}

func (as *ActionSuite) Test_TenantsResource_Create() {
	// var tenants = models.Tenants{}
	// //var tenat
	// res := as.JSON("/tenants").Post()
	// as.Equal(200, res.Code)
	// res.Bind(tenants)
	as.False(false)
}

func (as *ActionSuite) Test_TenantsResource_Update() {
	as.False(false)
}

func (as *ActionSuite) Test_TenantsResource_Destroy() {
	as.False(false)
}
