package actions

import (
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/nulls"
)

func (as *ActionSuite) Test_CarriersList() {
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
	newCarrier := as.createCarrier("carrier", models.CarrierTypeAir, nulls.Time{}, firmino.TenantID, firmino.ID)

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, "/carriers")
			res := req.Get()
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusOK {
				var carriers = models.Carriers{}
				res.Bind(&carriers)
				as.Equal(test.carrierCount, len(carriers))
				if test.carrierCount > 0 {
					as.Equal(newCarrier.Name, carriers[0].Name)
				}
			}
		})
	}
}

func (as *ActionSuite) Test_CarriersListFilter() {
	as.LoadFixture("Tenant bootstrap")
	var username = "firmino"
	user := as.getLoggedInUser(username)
	var prefixes = []string{"ਪੰਜਾਬੀ", "Test"}
	for _, p := range prefixes {
		for i := 0; i < 5; i++ {
			carrierType := models.CarrierTypeVessel
			if i%2 == 0 {
				carrierType = models.CarrierTypeAir
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
		as.Equal(models.CarrierTypeVessel.String(), v.Type)
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

func (as *ActionSuite) Test_CarriersListOrder() {
	as.LoadFixture("Tenant bootstrap")
	var username = "firmino"
	user := as.getLoggedInUser(username)
	var prefixes = []string{"ਪੰਜਾਬੀ", "Test"}
	var now = time.Now().UTC()
	for _, p := range prefixes {
		for i := 0; i < 3; i++ {
			carrierType := models.CarrierTypeVessel
			// Create eta based on the index. The smaller index should arrive closer to current time
			hours := i + 1
			if i%2 == 0 {
				carrierType = models.CarrierTypeAir
				// Setup some etas in the past
				hours = hours * -1
			}
			eta := now.Add(time.Duration(time.Hour * time.Duration(hours)))
			_ = as.createCarrier(fmt.Sprintf("%s-%d", p, i), carrierType, nulls.NewTime(eta), user.TenantID, user.ID)
		}
	}
	// Since ਪੰਜਾਬੀ are created first, they'll be farther from now than the corresponding Test based on created_at time
	var expectedOrder = []string{"Test-0", "ਪੰਜਾਬੀ-0", "Test-1", "ਪੰਜਾਬੀ-1", "Test-2", "ਪੰਜਾਬੀ-2"}
	req := as.setupRequest(user, "/carriers")
	res := req.Get()
	as.Equal(http.StatusOK, res.Code)
	var carriers = models.Carriers{}
	res.Bind(&carriers)
	as.Equal(6, len(carriers))

	for i, v := range carriers {
		as.Equal(expectedOrder[i], v.Name)
	}

}
func (as *ActionSuite) Test_CarriersShow() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
	}{
		{"mane", http.StatusOK},
		{"firmino", http.StatusOK},
		{"allan", http.StatusNotFound},
		{"salah", http.StatusOK},
		{"klopp", http.StatusOK},
		{"adidas", http.StatusOK},
	}
	lewin := as.getLoggedInUser("lewin")
	firmino := as.getLoggedInUser("firmino")
	as.NotEqual(firmino.TenantID, lewin.TenantID)
	eta := nulls.NewTime(time.Now())
	var carriers = []*models.Carrier{as.createCarrier("carr1", models.CarrierTypeAir, eta, firmino.TenantID, firmino.ID),
		as.createCarrier("carr2", models.CarrierTypeRail, eta, lewin.TenantID, lewin.ID)}
	as.NotEqual(carriers[0].TenantID, carriers[1].TenantID)

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			for _, v := range carriers {
				req := as.setupRequest(user, fmt.Sprintf("/carriers/%s", v.ID))
				res := req.Get()
				if v.TenantID == user.TenantID || user.IsSuperAdmin() {
					as.Equal(test.responseCode, res.Code)
				} else {
					as.Equal(http.StatusNotFound, res.Code)
				}
				if res.Code == http.StatusOK {
					var carrier = models.Carrier{}
					res.Bind(&carrier)
					as.Equal(v.Name, carrier.Name)
					as.Equal(v.Type, carrier.Type)
					as.Equal(eta.Time.UTC().Truncate(time.Minute), carrier.Eta.Time)
				}
			}
		})
	}
}

func (as *ActionSuite) Test_CarriersCreate() {
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
			carrierType := models.CarrierTypeAir
			if i%2 == 0 {
				carrierType = models.CarrierTypeRail
			}
			newCarrier := models.Carrier{Name: user.Username, Type: carrierType.String(), TenantID: firmino.TenantID}
			req := as.setupRequest(user, "/carriers")
			res := req.Post(newCarrier)
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusCreated {
				var carrier = models.Carrier{}
				res.Bind(&carrier)
				as.Equal(newCarrier.Name, carrier.Name)
				as.Equal(newCarrier.Type, carrier.Type)
				// should be null
				as.False(carrier.Eta.Valid)
				as.Equal(user.TenantID, carrier.TenantID)
			}
		})
	}
}

func (as *ActionSuite) Test_CarriersUpdate() {
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
			carrierType := models.CarrierTypeAir
			eta := nulls.NewTime(time.Now())
			newCarrier := as.createCarrier(user.Username, carrierType, eta, firmino.TenantID, firmino.ID)
			req := as.setupRequest(user, fmt.Sprintf("/carriers/%s", newCarrier.ID))
			// Try to update ID and tenant ID. Expect these calls to be excluded at update
			updatedCarrier := models.Carrier{Name: fmt.Sprintf("not%s", test.username), Type: models.CarrierTypeRail.String(), Eta: nulls.NewTime(eta.Time.Add(1)), ID: user.ID, TenantID: user.ID}
			res := req.Put(updatedCarrier)
			as.Equal(test.responseCode, res.Code)
			var dbCarrier = *newCarrier
			err := as.DB.Reload(&dbCarrier)
			as.Nil(err)
			if res.Code == http.StatusOK {
				var carrier = models.Carrier{}
				res.Bind(&carrier)
				as.Equal(updatedCarrier.Name, carrier.Name)
				as.Equal(updatedCarrier.Type, carrier.Type)
				as.Equal(updatedCarrier.Eta.Time.UTC().Truncate(time.Minute), carrier.Eta.Time)
				as.Equal(newCarrier.ID, carrier.ID)
				as.Equal(dbCarrier.Name, carrier.Name)
			} else {
				// Ensure update did not succeed
				as.Equal(dbCarrier.Name, newCarrier.Name)
			}
		})
	}
}

func (as *ActionSuite) Test_CarriersDestroy() {
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
			var name = fmt.Sprintf("carrier%s", test.username)
			newCarrier := &models.Carrier{Name: name, Type: models.CarrierTypeRail.String(), TenantID: firmino.TenantID, CreatedBy: firmino.ID}
			v, err := as.DB.ValidateAndCreate(newCarrier)
			as.Nil(err)
			as.Equal(0, len(v.Errors))

			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, fmt.Sprintf("/carriers/%s", newCarrier.ID))
			res := req.Delete()
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusNoContent {
				// Check if actually deleted
				carrier := models.Carrier{}
				err = as.DB.Where("name=?", name).First(&carrier)
				as.Equal(err, sql.ErrNoRows)
			} else {
				carrier := models.Carrier{}
				err = as.DB.Where("name=?", name).First(&carrier)
				//Not deleted yet
				as.Equal(name, carrier.Name)
			}
		})
	}
}
