package actions

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bigpanther/trober/models"
)

func (as *ActionSuite) Test_CarriersResource_List() {
	as.False(false)
}

func (as *ActionSuite) Test_CarriersResource_Show() {
	as.False(false)
}

func (as *ActionSuite) Test_CarriersResource_Create() {
	as.False(false)
}

func (as *ActionSuite) Test_CarriersResource_Update() {
	as.False(false)
}

func (as *ActionSuite) Test_CarriersResource_Destroy() {
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
			var name = fmt.Sprintf("carrier%s", test.username)
			newcarrier := &models.Carrier{Name: name, Type: "Port", TenantID: firmino.TenantID, CreatedBy: firmino.ID}
			v, err := as.DB.ValidateAndCreate(newcarrier)
			as.Nil(err)
			as.Equal(0, len(v.Errors))

			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, fmt.Sprintf("/carriers/%s", newcarrier.ID))
			res := req.Delete()
			as.Equal(test.responseCode, res.Code)
			if test.responseCode == http.StatusOK {
				var carrier = models.Carrier{}
				res.Bind(&carrier)
				as.Equal(name, carrier.Name)
				// Check if actually deleted
				carrier = models.Carrier{}
				err = as.DB.Where("name=?", name).First(&carrier)
				as.Contains(err.Error(), "no rows in result set")
			} else {
				carrier := models.Carrier{}
				err = as.DB.Where("name=?", name).First(&carrier)
				//Not deleted yet
				as.Equal(name, carrier.Name)
			}
		})
	}
}
