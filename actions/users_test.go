package actions

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bigpanther/trober/models"
)

func (as *ActionSuite) Test_UsersResource_List() {
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
			if test.responseCode == http.StatusOK {
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

func (as *ActionSuite) Test_UsersResource_List_Filter() {
	as.LoadFixture("Tenant bootstrap")
	var username = "firmino"
	user := as.getLoggedInUser(username)
	var prefixes = []string{"ਪੰਜਾਬੀ", "Test"}
	for _, p := range prefixes {
		for i := 0; i < 5; i++ {
			userRole := "Driver"
			if i%2 == 0 {
				userRole = "BackOffice"
			}
			newUser := &models.User{Name: fmt.Sprintf("%s - %d", p, i), Role: userRole, Username: fmt.Sprintf("%s-%d", p, i), Email: fmt.Sprintf("%s-%d@bigpanther.ca", p, i), TenantID: user.TenantID}
			v, err := as.DB.ValidateAndCreate(newUser)
			as.Nil(err)
			as.Equal(0, len(v.Errors))
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

	as.True(klopp.IsSuperAdmin())
	req = as.setupRequest(klopp, fmt.Sprintf("/users?tenant_id=%s", lewin.TenantID))
	res = req.Get()
	as.Equal(http.StatusOK, res.Code)
	users = models.Users{}
	res.Bind(&users)
	as.Equal(5, len(users))

}
func (as *ActionSuite) Test_UsersResource_Show() {
	as.False(false)
}

func (as *ActionSuite) Test_UsersResource_Create() {
	as.False(false)
}

func (as *ActionSuite) Test_UsersResource_Update() {
	as.False(false)
}

func (as *ActionSuite) Test_UsersResource_Destroy() {
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
			newUser := &models.User{Name: name, Role: "Driver", Username: name, Email: fmt.Sprintf("user%s@bigpanther.ca", test.username), TenantID: firmino.TenantID}
			v, err := as.DB.ValidateAndCreate(newUser)
			as.Nil(err)
			as.Equal(0, len(v.Errors))

			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, fmt.Sprintf("/users/%s", newUser.ID))
			res := req.Delete()
			as.Equal(test.responseCode, res.Code)
			if test.responseCode == http.StatusOK {
				var user = models.User{}
				res.Bind(&user)
				as.Equal(name, user.Name)
				// Check if actually deleted
				user = models.User{}
				err = as.DB.Where("name=?", name).First(&user)
				as.Contains(err.Error(), "no rows in result set")
			} else {
				user := models.User{}
				err = as.DB.Where("name=?", name).First(&user)
				//Not deleted yet
				as.Equal(name, user.Name)
			}
		})
	}
}
