package actions

import (
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/nulls"
)

func (as *ActionSuite) Test_UsersList() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
		userCount    int
		sameTenant   bool
	}{
		{"mane", http.StatusOK, 5, true},
		{"firmino", http.StatusOK, 5, true},
		{"allan", http.StatusNotFound, 0, true},
		{"salah", http.StatusNotFound, 0, true},
		{"klopp", http.StatusOK, 11, false},
		{"adidas", http.StatusNotFound, 0, true},
	}

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, "/users")
			res := req.Get()
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusOK {
				var users = models.Users{}
				res.Bind(&users)
				as.Equal(test.userCount, len(users))
				if test.sameTenant {
					for _, u := range users {
						as.Equal(user.TenantID, u.TenantID)
					}
				}
			}
		})
	}
}

func (as *ActionSuite) Test_UsersListFilter() {
	as.LoadFixture("Tenant bootstrap")
	var username = "firmino"
	user := as.getLoggedInUser(username)
	var prefixes = []string{"ਪੰਜਾਬੀ", "Test"}
	for _, p := range prefixes {
		for i := 0; i < 5; i++ {
			userRole := models.UserRoleDriver
			if i%2 == 0 {
				userRole = models.UserRoleBackOffice
			}
			_ = as.createUser(fmt.Sprintf("%s-%d", p, i), userRole, fmt.Sprintf("%s-%d@bigpanther.ca", p, i), user.TenantID, nulls.UUID{})
		}
	}
	req := as.setupRequest(user, "/users?name=ਪੰ&role=Driver")
	res := req.Get()
	as.Equal(http.StatusOK, res.Code)
	var users = models.Users{}
	res.Bind(&users)
	as.Equal(2, len(users))
	for _, v := range users {
		as.Contains(v.Name, "ਪੰਜਾਬੀ")
		as.Equal("Driver", v.Role)
	}
	klopp := as.getLoggedInUser("klopp")
	as.NotEqual(klopp.TenantID, user.TenantID)

	as.False(user.IsSuperAdmin())
	req = as.setupRequest(user, fmt.Sprintf("/users?tenant_id=%s", klopp.TenantID))
	res = req.Get()
	as.Equal(http.StatusOK, res.Code)
	users = models.Users{}
	res.Bind(&users)
	as.Equal(0, len(users))

	lewin := as.getLoggedInUser("lewin")
	as.NotEqual(klopp.TenantID, lewin.TenantID)
	as.NotEqual(lewin.TenantID, user.TenantID)

	as.True(klopp.IsSuperAdmin())
	req = as.setupRequest(klopp, fmt.Sprintf("/users?tenant_id=%s", lewin.TenantID))
	res = req.Get()
	as.Equal(http.StatusOK, res.Code)
	users = models.Users{}
	res.Bind(&users)
	as.Equal(5, len(users))

}
func (as *ActionSuite) Test_UsersShow() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
		otherUser    string
	}{
		{"mane", http.StatusOK, "firmino"},
		{"mane", http.StatusNotFound, "allan"},
		{"mane", http.StatusNotFound, "klopp"},
		{"firmino", http.StatusOK, "salah"},
		{"firmino", http.StatusNotFound, "klopp"},
		{"salah", http.StatusNotFound, "firmino"},
		{"salah", http.StatusNotFound, "allan"},
		{"salah", http.StatusNotFound, "klopp"},
		{"allan", http.StatusNotFound, "lewin"},
		{"allan", http.StatusNotFound, "mane"},
		{"allan", http.StatusNotFound, "klopp"},
		{"klopp", http.StatusOK, "mane"},
		{"klopp", http.StatusOK, "allan"},
		{"klopp", http.StatusOK, "adidas"},
		{"adidas", http.StatusNotFound, "mane"},
		{"adidas", http.StatusNotFound, "klopp"},
		{"adidas", http.StatusNotFound, "lewin"},
	}
	for _, test := range tests {
		as.T().Run(fmt.Sprint(test.username, test.otherUser), func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			otherUser := as.getLoggedInUser(test.otherUser)
			var UserIDs = map[string]string{
				user.Username:      user.ID.String(),
				otherUser.Username: otherUser.ID.String(),
			}
			for k, v := range UserIDs {
				req := as.setupRequest(user, fmt.Sprintf("/users/%s", v))
				res := req.Get()
				if !user.IsAtLeastBackOffice() {
					as.Equal(http.StatusNotFound, res.Code)
				}
				if k == test.otherUser {
					as.Equal(test.responseCode, res.Code)
				}
				if res.Code == http.StatusOK {
					var user = models.User{}
					res.Bind(&user)
					as.Equal(k, user.Username)
				}
			}
		})
	}
}

func (as *ActionSuite) Test_UsersCreate() {
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
	for i, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			userRole := models.UserRoleBackOffice
			if i%2 == 0 {
				userRole = models.UserRoleDriver
			}
			newUser := models.User{Name: test.username, Username: "irrelevant", Email: fmt.Sprintf("%stest@bigpanther.ca", test.username), Role: userRole.String(), TenantID: firmino.TenantID}
			req := as.setupRequest(user, "/users")
			res := req.Post(newUser)
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusCreated {
				var u = models.User{}
				res.Bind(&u)
				as.Equal(newUser.Name, u.Name)
				as.Equal(newUser.Role, u.Role)
				as.Equal(newUser.CustomerID, nulls.UUID{})
				if user.IsSuperAdmin() {
					as.Equal(firmino.TenantID, u.TenantID)
				} else {
					as.Equal(user.TenantID, u.TenantID)
				}
				as.Contains(u.Username, "invited-")
				// try creating superadmin
				newUser = models.User{Name: test.username, Username: "irrelevant", Email: fmt.Sprintf("%ssuperadmintest@bigpanther.ca", test.username), Role: models.UserRoleSuperAdmin.String(), TenantID: firmino.TenantID}
				req := as.setupRequest(user, "/users")
				res := req.Post(newUser)
				as.Equal(http.StatusBadRequest, res.Code)
			}
		})
	}
}
func (as *ActionSuite) Test_UsersCreateCustomerRole() {
	as.LoadFixture("Tenant bootstrap")
	customer := as.getCustomer("UEFA Liv")
	var tests = []struct {
		username     string
		customerID   nulls.UUID
		responseCode int
	}{
		{"firmino", nulls.UUID{}, http.StatusBadRequest},
		{"firmino", nulls.NewUUID(customer.ID), http.StatusCreated},
		{"klopp", nulls.NewUUID(customer.ID), http.StatusCreated},
		{"rodriguez", nulls.NewUUID(customer.ID), http.StatusBadRequest},
	}
	var firmino = as.getLoggedInUser("firmino")
	for i, test := range tests {
		as.T().Run(fmt.Sprintf("%s%d", test.username, i), func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			userRole := models.UserRoleCustomer
			newUser := models.User{Name: test.username, Username: "irrelevant", Email: fmt.Sprintf("%stest@bigpanther.ca", test.username), Role: userRole.String(), TenantID: firmino.TenantID, CustomerID: test.customerID}
			req := as.setupRequest(user, "/users")
			res := req.Post(newUser)
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusCreated {
				var u = models.User{}
				res.Bind(&u)
				as.Equal(newUser.Name, u.Name)
				as.Equal(models.UserRoleCustomer.String(), u.Role)
				as.Equal(newUser.CustomerID, u.CustomerID)
				if user.IsSuperAdmin() {
					as.Equal(firmino.TenantID, u.TenantID)
				} else {
					as.Equal(user.TenantID, u.TenantID)
				}
				as.Contains(u.Username, "invited-")
				// try creating superadmin
				newUser = models.User{Name: test.username, Username: "irrelevant", Email: fmt.Sprintf("%ssuperadmintest@bigpanther.ca", test.username), Role: models.UserRoleSuperAdmin.String(), TenantID: firmino.TenantID}
				req := as.setupRequest(user, "/users")
				res := req.Post(newUser)
				as.Equal(http.StatusBadRequest, res.Code)
			}
		})
	}
}

func (as *ActionSuite) Test_UsersUpdate() {
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
			role := models.UserRoleDriver
			newUser := as.createUser(fmt.Sprintf("test%s", user.Username), role, fmt.Sprintf("test%s@bigpanther.ca", test.username), firmino.TenantID, nulls.UUID{})
			req := as.setupRequest(user, fmt.Sprintf("/users/%s", newUser.ID))
			// Try to update ID and tenant ID. Expect these calls to be excluded at update
			updatedUser := models.User{Name: fmt.Sprintf("not%s", test.username), Role: models.UserRoleAdmin.String(), ID: user.ID, TenantID: user.ID}
			res := req.Put(updatedUser)
			as.Equal(test.responseCode, res.Code)
			var dbUser = *newUser
			err := as.DB.Reload(&dbUser)
			as.Nil(err)
			if res.Code == http.StatusOK {
				var user = models.User{}
				res.Bind(&user)
				as.Equal(updatedUser.Name, user.Name)
				as.Equal(updatedUser.Role, user.Role)
				as.Equal(newUser.ID, user.ID)
				as.Equal(dbUser.Name, user.Name)
			} else {
				// Ensure update did not succeed
				as.Equal(dbUser.Name, newUser.Name)
			}
		})
	}
}

func (as *ActionSuite) Test_UsersDestroy() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
	}{
		{"klopp", http.StatusOK},
		{"firmino", http.StatusOK},
		{"mane", http.StatusOK},
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
			var name = fmt.Sprintf("user%s", test.username)
			newUser := as.createUser(name, models.UserRoleDriver, fmt.Sprintf("user%s@bigpanther.ca", test.username), firmino.TenantID, nulls.UUID{})
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, fmt.Sprintf("/users/%s", newUser.ID))
			res := req.Delete()
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusOK {
				var user = models.User{}
				res.Bind(&user)
				as.Equal(name, user.Name)
				// Check if actually deleted
				user = models.User{}
				var err = as.DB.Where("name=?", name).First(&user)
				as.Equal(err, sql.ErrNoRows)
			} else {
				user := models.User{}
				err := as.DB.Where("name=?", name).First(&user)
				as.Nil(err)
				//Not deleted yet
				as.Equal(name, user.Name)
			}
		})
	}
}
