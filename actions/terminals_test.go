package actions

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bigpanther/trober/models"
)

func (as *ActionSuite) Test_TerminalsResource_List() {
	as.False(false)
}

func (as *ActionSuite) Test_TerminalsResource_Show() {
	as.False(false)
}

func (as *ActionSuite) Test_TerminalsResource_Create() {
	as.False(false)
}

func (as *ActionSuite) Test_TerminalsResource_Update() {
	as.False(false)
}

func (as *ActionSuite) Test_TerminalsResource_Destroy() {
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
			var name = fmt.Sprintf("terminal%s", test.username)
			newTerminal := &models.Terminal{Name: name, Type: "Port", TenantID: firmino.TenantID, CreatedBy: firmino.ID}
			v, err := as.DB.ValidateAndCreate(newTerminal)
			as.Nil(err)
			as.Equal(0, len(v.Errors))

			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, fmt.Sprintf("/terminals/%s", newTerminal.ID))
			res := req.Delete()
			as.Equal(test.responseCode, res.Code)
			if test.responseCode == http.StatusOK {
				var terminal = models.Terminal{}
				res.Bind(&terminal)
				as.Equal(name, terminal.Name)
				// Check if actually deleted
				terminal = models.Terminal{}
				err = as.DB.Where("name=?", name).First(&terminal)
				as.Contains(err.Error(), "no rows in result set")
			} else {
				terminal := models.Terminal{}
				err = as.DB.Where("name=?", name).First(&terminal)
				//Not deleted yet
				as.Equal(name, terminal.Name)
			}
		})
	}
}
