package actions

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/nulls"
)

func (as *ActionSuite) Test_TenantsResource_List() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
		tenantCount  int
	}{
		{"mane", http.StatusNotFound, 0},
		{"firmino", http.StatusNotFound, 0},
		{"allan", http.StatusNotFound, 0},
		{"salah", http.StatusNotFound, 0},
		{"klopp", http.StatusOK, 3},
		{"adidas", http.StatusNotFound, 0},
	}

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, "/tenants")
			res := req.Get()
			as.Equal(test.responseCode, res.Code)
			if test.responseCode == http.StatusOK {
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
		{"mane", http.StatusOK, "Big Panther Liverpool"},
		{"firmino", http.StatusOK, "Big Panther Liverpool"},
		{"allan", http.StatusNotFound, "Big Panther Everton"},
		{"salah", http.StatusOK, "Big Panther Liverpool"},
		{"klopp", http.StatusOK, "Big Panther Test"},
		{"adidas", http.StatusOK, "Big Panther Everton"},
	}

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, fmt.Sprintf("/tenants/%s", user.TenantID))
			res := req.Get()
			as.Equal(test.responseCode, res.Code)
			if test.responseCode == http.StatusOK {
				var tenant = &models.Tenant{}
				res.Bind(tenant)
				as.Equal(test.tenantName, tenant.Name)
			}
		})
	}
}

func (as *ActionSuite) Test_TenantsResource_Create() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
	}{
		{"mane", http.StatusNotFound},
		{"firmino", http.StatusNotFound},
		{"allan", http.StatusNotFound},
		{"salah", http.StatusNotFound},
		{"adidas", http.StatusNotFound},
		{"klopp", http.StatusCreated},
	}
	newTenant := &models.Tenant{Name: "Test", Type: "Production", Code: nulls.NewString("someC")}

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, "/tenants")
			res := req.Post(newTenant)
			as.Equal(test.responseCode, res.Code)
			if test.responseCode == http.StatusCreated {
				var tenant = &models.Tenant{}
				res.Bind(tenant)
				as.Equal("Test", tenant.Name)
				tenant = &models.Tenant{}
				var err = as.DB.Where("name=?", "Test").First(tenant)
				as.Nil(err)
				as.Equal("Test", tenant.Name)
			}
		})
	}
}

func (as *ActionSuite) Test_TenantsResource_Update() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
	}{
		{"mane", http.StatusNotFound},
		{"firmino", http.StatusNotFound},
		{"allan", http.StatusNotFound},
		{"salah", http.StatusNotFound},
		{"adidas", http.StatusNotFound},
		{"klopp", http.StatusOK},
	}
	newTenant := &models.Tenant{Name: "Test", Type: "Production", Code: nulls.NewString("someC")}
	v, err := as.DB.ValidateAndCreate(newTenant)
	as.Nil(err)
	as.Equal(0, len(v.Errors))

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, fmt.Sprintf("/tenants/%s", newTenant.ID))
			newTenant.Name = "New Test"
			res := req.Put(newTenant)
			as.Equal(test.responseCode, res.Code)
			if test.responseCode == http.StatusOK {
				var tenant = &models.Tenant{}
				res.Bind(tenant)
				as.Equal("New Test", tenant.Name)
				// Check if actually updated
				tenant = &models.Tenant{}
				err = as.DB.Where("name=?", "New Test").First(tenant)
				as.Equal("New Test", tenant.Name)
			} else {
				tenant := &models.Tenant{}
				err = as.DB.Where("name=?", "Test").First(tenant)
				//Not updated yet
				as.Equal("Test", tenant.Name)
			}
		})
	}
}

func (as *ActionSuite) Test_TenantsResource_Destroy() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
	}{
		{"mane", http.StatusNotFound},
		{"firmino", http.StatusNotFound},
		{"allan", http.StatusNotFound},
		{"salah", http.StatusNotFound},
		{"adidas", http.StatusNotFound},
		{"klopp", http.StatusOK},
	}
	newTenant := &models.Tenant{Name: "Test", Type: "Production", Code: nulls.NewString("someC")}
	v, err := as.DB.ValidateAndCreate(newTenant)
	as.Nil(err)
	as.Equal(0, len(v.Errors))

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, fmt.Sprintf("/tenants/%s", newTenant.ID))
			res := req.Delete()
			as.Equal(test.responseCode, res.Code)
			if test.responseCode == http.StatusOK {
				var tenant = &models.Tenant{}
				res.Bind(tenant)
				as.Equal("Test", tenant.Name)
				// Check if actually deleted
				tenant = &models.Tenant{}
				err = as.DB.Where("name=?", "Test").First(tenant)
				as.Contains(err.Error(), "no rows in result set")
			} else {
				tenant := &models.Tenant{}
				err = as.DB.Where("name=?", "Test").First(tenant)
				//Not deleted yet
				as.Equal("Test", tenant.Name)
			}
		})
	}
}
