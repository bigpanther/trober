package actions

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/nulls"
)

func (as *ActionSuite) Test_CarriersResource_List() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
		carrierCount int
	}{
		{"klopp", http.StatusOK, 1},
		{"firmino", http.StatusOK, 1},
		{"mane", http.StatusOK, 1},
		{"salah", http.StatusOK, 1},
		{"nike", http.StatusOK, 1},
		{"coutinho", http.StatusNotFound, 0},
		{"richarlson", http.StatusOK, 0},
		{"rodriguez", http.StatusOK, 0},
		{"lewin", http.StatusOK, 0},
		{"allan", http.StatusNotFound, 0},
		{"adidas", http.StatusOK, 0},
	}
	firmino := as.getLoggedInUser("firmino")
	newcarrier := as.createCarrier("carrier", "Port", nulls.Time{}, firmino.TenantID, firmino.ID)

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, "/carriers")
			res := req.Get()
			as.Equal(test.responseCode, res.Code)
			if test.responseCode == http.StatusOK {
				var carriers = models.Carriers{}
				res.Bind(&carriers)
				as.Equal(test.carrierCount, len(carriers))
				if test.carrierCount > 0 {
					as.Equal(newcarrier.Name, carriers[0].Name)
				}
			}
		})
	}
}

func (as *ActionSuite) Test_CarriersResource_List_Filter() {
	as.LoadFixture("Tenant bootstrap")
	var username = "firmino"
	user := as.getLoggedInUser(username)
	var prefixes = []string{"ਪੰਜਾਬੀ", "Test"}
	for _, p := range prefixes {
		for i := 0; i < 5; i++ {
			carrierType := "Vessel"
			if i%2 == 0 {
				carrierType = "Air"
			}
			_ = as.createCarrier(fmt.Sprintf("%s-%d", p, i), carrierType, nulls.Time{}, user.TenantID, user.ID)
		}
	}
	req := as.setupRequest(user, "/carriers?name=ਪੰ&type=Vessel")
	res := req.Get()
	as.Equal(http.StatusOK, res.Code)
	var carriers = models.Carriers{}
	res.Bind(&carriers)
	as.Equal(2, len(carriers))
	for _, v := range carriers {
		as.Contains(v.Name, "ਪੰਜਾਬੀ")
		as.Equal("Vessel", v.Type)
	}
	klopp := as.getLoggedInUser("klopp")
	as.NotEqual(klopp.TenantID, user.TenantID)

	as.False(user.IsSuperAdmin())
	req = as.setupRequest(user, fmt.Sprintf("/carriers?tenant_id=%s", klopp.TenantID))
	res = req.Get()
	as.Equal(http.StatusOK, res.Code)
	carriers = models.Carriers{}
	res.Bind(&carriers)
	as.Equal(0, len(carriers))

	lewin := as.getLoggedInUser("lewin")
	as.NotEqual(klopp.TenantID, lewin.TenantID)
	as.NotEqual(lewin.TenantID, user.TenantID)
	as.True(klopp.IsSuperAdmin())

	req = as.setupRequest(klopp, fmt.Sprintf("/carriers?tenant_id=%s", lewin.TenantID))
	res = req.Get()
	as.Equal(http.StatusOK, res.Code)
	carriers = models.Carriers{}
	res.Bind(&carriers)
	as.Equal(0, len(carriers))
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
