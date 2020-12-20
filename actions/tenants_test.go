package actions

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/nulls"
)

func (as *ActionSuite) Test_TenantsList() {
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
				var tenants = models.Tenants{}
				res.Bind(&tenants)
				as.Equal(test.tenantCount, len(tenants))
			}
		})
	}
}
func (as *ActionSuite) Test_TenantsListOrder() {
	as.LoadFixture("Tenant bootstrap")
	var username = "klopp"
	newTenant := &models.Tenant{Name: "Test", Type: "Production", Code: nulls.NewString("someC")}
	v, err := as.DB.ValidateAndCreate(newTenant)
	as.Nil(err)
	as.Equal(0, len(v.Errors))
	user := as.getLoggedInUser(username)
	req := as.setupRequest(user, "/tenants")
	res := req.Get()
	as.Equal(http.StatusOK, res.Code)
	var tenants = models.Tenants{}
	res.Bind(&tenants)
	as.Equal(len(tenants), 4)
	as.Equal("Test", (tenants)[0].Name)
}

func (as *ActionSuite) Test_TenantsListPagination() {
	as.LoadFixture("Tenant bootstrap")
	var username = "klopp"
	newTenant := &models.Tenant{Name: "Test", Type: "Production", Code: nulls.NewString("someC")}
	v, err := as.DB.ValidateAndCreate(newTenant)
	as.Nil(err)
	as.Equal(0, len(v.Errors))
	user := as.getLoggedInUser(username)
	req := as.setupRequest(user, "/tenants?page=1&per_page=1")
	res := req.Get()
	as.Equal(http.StatusOK, res.Code)
	var tenants = models.Tenants{}
	res.Bind(&tenants)
	as.Equal(len(tenants), 1)
	as.Equal("Test", (tenants)[0].Name)
	req = as.setupRequest(user, "/tenants?page=2&per_page=2")
	res = req.Get()
	as.Equal(http.StatusOK, res.Code)
	tenants = models.Tenants{}
	res.Bind(&tenants)
	as.Equal(len(tenants), 2)
	for _, v := range tenants {
		as.Contains(v.Name, "Big Panther")
	}
}
func (as *ActionSuite) Test_TenantsListFilter() {
	as.LoadFixture("Tenant bootstrap")
	var prefixes = []string{"ਪੰਜਾਬੀ", "Test"}
	for _, p := range prefixes {
		for i := 0; i < 5; i++ {
			tenantType := "Test"
			if i%2 == 0 {
				tenantType = "Production"
			}
			newTenant := &models.Tenant{Name: fmt.Sprintf("%s - %d", p, i), Type: tenantType, Code: nulls.NewString("someC")}
			v, err := as.DB.ValidateAndCreate(newTenant)
			as.Nil(err)
			as.Equal(0, len(v.Errors))
		}
	}

	var username = "klopp"
	user := as.getLoggedInUser(username)
	req := as.setupRequest(user, "/tenants?name=ਪੰ")
	res := req.Get()
	as.Equal(http.StatusOK, res.Code)
	var tenants = models.Tenants{}
	res.Bind(&tenants)
	as.Equal(5, len(tenants))
	for _, v := range tenants {
		as.Contains(v.Name, "ਪੰਜਾਬੀ")
	}
	req = as.setupRequest(user, "/tenants?name=tes&type=Production")
	res = req.Get()
	as.Equal(http.StatusOK, res.Code)
	tenants = models.Tenants{}
	res.Bind(&tenants)
	as.Equal(3, len(tenants))
	for _, v := range tenants {
		as.Contains(v.Name, "Test")
	}
}

func (as *ActionSuite) Test_TenantsShow() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
		tenantName   string
	}{
		{"mane", http.StatusNotFound, "Big Panther Liverpool"},
		{"firmino", http.StatusNotFound, "Big Panther Liverpool"},
		{"allan", http.StatusNotFound, "Big Panther Everton"},
		{"salah", http.StatusNotFound, "Big Panther Liverpool"},
		{"klopp", http.StatusOK, "Big Panther Test"},
		{"adidas", http.StatusNotFound, "Big Panther Everton"},
	}
	firmino := as.getLoggedInUser("firmino")
	klopp := as.getLoggedInUser("klopp")
	as.NotEqual(firmino.TenantID, klopp.TenantID)
	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			var tenantIDs = map[string]string{
				"Big Panther Test":      klopp.TenantID.String(),
				"Big Panther Liverpool": firmino.TenantID.String(),
				test.tenantName:         user.TenantID.String(),
			}
			for k, v := range tenantIDs {
				req := as.setupRequest(user, fmt.Sprintf("/tenants/%s", v))
				res := req.Get()
				as.Equal(test.responseCode, res.Code)
				if test.responseCode == http.StatusOK {
					var tenant = models.Tenant{}
					res.Bind(&tenant)
					as.Equal(k, tenant.Name)
				}
			}
		})
	}
}

func (as *ActionSuite) Test_TenantsCreate() {
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
				var tenant = models.Tenant{}
				res.Bind(&tenant)
				as.Equal("Test", tenant.Name)
				tenant = models.Tenant{}
				var err = as.DB.Where("name=?", "Test").First(&tenant)
				as.Nil(err)
				as.Equal("Test", tenant.Name)
			}
		})
	}
}

func (as *ActionSuite) Test_TenantsUpdate() {
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
				var tenant = models.Tenant{}
				res.Bind(&tenant)
				as.Equal("New Test", tenant.Name)
				// Check if actually updated
				tenant = models.Tenant{}
				err = as.DB.Where("name=?", "New Test").First(&tenant)
				as.Equal("New Test", tenant.Name)
			} else {
				tenant := models.Tenant{}
				err = as.DB.Where("name=?", "Test").First(&tenant)
				//Not updated yet
				as.Equal("Test", tenant.Name)
			}
		})
	}
}

func (as *ActionSuite) Test_TenantsDestroy() {
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
				var tenant = models.Tenant{}
				res.Bind(&tenant)
				as.Equal("Test", tenant.Name)
				// Check if actually deleted
				tenant = models.Tenant{}
				err = as.DB.Where("name=?", "Test").First(&tenant)
				as.Contains(err.Error(), "no rows in result set")
			} else {
				tenant := models.Tenant{}
				err := as.DB.Where("name=?", "Test").First(&tenant)
				//Not deleted yet
				as.Nil(err)
				as.Equal("Test", tenant.Name)
			}
		})
	}
}
