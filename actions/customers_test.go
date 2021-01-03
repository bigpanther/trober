package actions

import (
	"database/sql"
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
			if res.Code == http.StatusOK {
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
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
	}{
		{"mane", http.StatusOK},
		{"firmino", http.StatusOK},
		{"allan", http.StatusNotFound},
		{"salah", http.StatusNotFound},
		{"klopp", http.StatusOK},
		{"adidas", http.StatusNotFound},
		{"nike", http.StatusNotFound},
		{"new", http.StatusOK},
	}
	lewin := as.getLoggedInUser("lewin")
	firmino := as.getLoggedInUser("firmino")
	as.NotEqual(firmino.TenantID, lewin.TenantID)
	var customers = []*models.Customer{as.createCustomer("cust1", firmino.TenantID, nulls.NewUUID(firmino.ID)),
		as.createCustomer("cust2", lewin.TenantID, nulls.NewUUID(lewin.ID))}
	// Create new user for customerID
	as.createUser("new", models.UserRoleCustomer, "new@bigpanther.ca", firmino.TenantID, nulls.NewUUID(customers[0].ID))
	as.NotEqual(customers[0].TenantID, customers[1].TenantID)

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			for _, v := range customers {
				req := as.setupRequest(user, fmt.Sprintf("/customers/%s", v.ID))
				res := req.Get()
				if v.TenantID == user.TenantID || user.IsSuperAdmin() {
					as.Equal(test.responseCode, res.Code)
				} else {
					as.Equal(http.StatusNotFound, res.Code)
				}
				if res.Code == http.StatusOK {
					var customer = models.Customer{}
					res.Bind(&customer)
					as.Equal(v.Name, customer.Name)
				}
			}
		})
	}

}

func (as *ActionSuite) Test_CustomersCreate() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
	}{
		{"mane", http.StatusCreated},
		{"firmino", http.StatusCreated},
		{"rodriguez", http.StatusCreated},
		{"allan", http.StatusNotFound},
		{"salah", http.StatusNotFound},
		{"klopp", http.StatusCreated},
		{"adidas", http.StatusNotFound},
	}
	var firmino = as.getLoggedInUser("firmino")
	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			newCustomer := models.Customer{Name: user.Username, TenantID: firmino.TenantID}
			req := as.setupRequest(user, "/customers")
			res := req.Post(newCustomer)
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusCreated {
				var customer = models.Customer{}
				res.Bind(&customer)
				as.Equal(newCustomer.Name, customer.Name)
				as.Equal(user.TenantID, customer.TenantID)
			}
		})
	}
}

func (as *ActionSuite) Test_CustomersUpdate() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
	}{
		{"mane", http.StatusOK},
		{"firmino", http.StatusOK},
		{"rodriguez", http.StatusNotFound},
		{"coutinho", http.StatusNotFound},
		{"allan", http.StatusNotFound},
		{"salah", http.StatusNotFound},
		{"klopp", http.StatusOK},
		{"adidas", http.StatusNotFound},
		{"nike", http.StatusNotFound},
	}
	var firmino = as.getLoggedInUser("firmino")
	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			newCustomer := as.createCustomer(user.Username, firmino.TenantID, nulls.NewUUID(firmino.ID))
			req := as.setupRequest(user, fmt.Sprintf("/customers/%s", newCustomer.ID))
			// Try to update ID and tenant ID. Expect these calls to be excluded at update
			updatedCustomer := models.Customer{Name: fmt.Sprintf("not%s", test.username), ID: user.ID, TenantID: user.ID}
			res := req.Put(updatedCustomer)
			as.Equal(test.responseCode, res.Code)
			var dbCustomer = *newCustomer
			err := as.DB.Reload(&dbCustomer)
			as.Nil(err)
			if res.Code == http.StatusOK {
				var customer = models.Customer{}
				res.Bind(&customer)
				as.Equal(updatedCustomer.Name, customer.Name)
				as.Equal(newCustomer.ID, customer.ID)
				as.Equal(dbCustomer.Name, customer.Name)
			} else {
				// Ensure update did not succeed
				as.Equal(dbCustomer.Name, newCustomer.Name)
			}
		})
	}
}

func (as *ActionSuite) Test_CustomersDestroy() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
	}{
		{"klopp", http.StatusNoContent},
		{"firmino", http.StatusNoContent},
		{"mane", http.StatusNoContent},
		{"salah", http.StatusNotFound},
		{"nike", http.StatusNotFound},
		{"coutinho", http.StatusNotFound},
		{"richarlson", http.StatusNotFound},
		{"rodriguez", http.StatusNotFound},
		{"lewin", http.StatusNotFound},
		{"allan", http.StatusNotFound},
		{"adidas", http.StatusNotFound},
	}
	firmino := as.getLoggedInUser("firmino")

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			var name = fmt.Sprintf("customer%s", test.username)
			newCustomer := &models.Customer{Name: name, TenantID: firmino.TenantID, CreatedBy: nulls.NewUUID(firmino.ID)}
			v, err := as.DB.ValidateAndCreate(newCustomer)
			as.Nil(err)
			as.Equal(0, len(v.Errors))

			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, fmt.Sprintf("/customers/%s", newCustomer.ID))
			res := req.Delete()
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusNoContent {
				// Check if actually deleted
				customer := models.Customer{}
				err = as.DB.Where("name=?", name).First(&customer)
				as.Equal(err, sql.ErrNoRows)
			} else {
				customer := models.Customer{}
				err = as.DB.Where("name=?", name).First(&customer)
				//Not deleted yet
				as.Equal(name, customer.Name)
			}
		})
	}
}
