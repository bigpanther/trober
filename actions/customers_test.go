package actions

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/nulls"
)

func (as *ActionSuite) Test_CustomersList() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username      string
		responseCode  int
		customerCount int
	}{
		{"klopp", http.StatusOK, 6},
		{"firmino", http.StatusOK, 3},
		{"mane", http.StatusOK, 3},
		{"salah", http.StatusNotFound, 0},
		{"nike", http.StatusNotFound, 0},
		{"coutinho", http.StatusNotFound, 0},
		{"richarlson", http.StatusOK, 3},
		{"rodriguez", http.StatusOK, 3},
		{"lewin", http.StatusNotFound, 0},
		{"allan", http.StatusNotFound, 0},
		{"adidas", http.StatusNotFound, 0},
	}
	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, "/customers")
			res := req.Get()
			as.Equal(test.responseCode, res.Code)
			if test.responseCode == http.StatusOK {
				var customers = models.Customers{}
				res.Bind(&customers)
				as.Equal(test.customerCount, len(customers))
				if test.username != "klopp" {
					for _, v := range customers {
						as.Equal(user.TenantID, v.TenantID)
					}
				}
			}
		})
	}
}

func (as *ActionSuite) Test_CustomersListFilter() {
	as.LoadFixture("Tenant bootstrap")
	var username = "firmino"
	user := as.getLoggedInUser(username)
	var prefixes = []string{"ਪੰਜਾਬੀ", "Test"}
	for _, p := range prefixes {
		_ = as.createCustomer(p, user.TenantID, nulls.NewUUID(user.ID))

	}
	req := as.setupRequest(user, "/customers?name=ਪੰ")
	res := req.Get()
	as.Equal(http.StatusOK, res.Code)
	var customers = models.Customers{}
	res.Bind(&customers)
	as.Equal(1, len(customers))
	for _, v := range customers {
		as.Equal(v.Name, "ਪੰਜਾਬੀ")
	}
	klopp := as.getLoggedInUser("klopp")
	as.NotEqual(klopp.TenantID, user.TenantID)

	as.False(user.IsSuperAdmin())
	req = as.setupRequest(user, fmt.Sprintf("/customers?tenant_id=%s", klopp.TenantID))
	res = req.Get()
	as.Equal(http.StatusOK, res.Code)
	customers = models.Customers{}
	res.Bind(&customers)
	as.Equal(0, len(customers))

	lewin := as.getLoggedInUser("lewin")
	as.NotEqual(klopp.TenantID, lewin.TenantID)
	as.NotEqual(lewin.TenantID, user.TenantID)
	as.True(klopp.IsSuperAdmin())

	req = as.setupRequest(klopp, fmt.Sprintf("/customers?tenant_id=%s", lewin.TenantID))
	res = req.Get()
	as.Equal(http.StatusOK, res.Code)
	customers = models.Customers{}
	res.Bind(&customers)
	as.Equal(3, len(customers))
}

func (as *ActionSuite) Test_CustomersShow() {
	as.False(false)
}

func (as *ActionSuite) Test_CustomersCreate() {
	as.False(false)
}

func (as *ActionSuite) Test_CustomersUpdate() {
	as.False(false)
}

func (as *ActionSuite) Test_CustomersDestroy() {
	as.False(false)
}
