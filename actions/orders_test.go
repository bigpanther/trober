package actions

import (
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"github.com/bigpanther/trober/models"
)

func (as *ActionSuite) Test_OrdersList() {
	as.False(false)
}

func (as *ActionSuite) Test_OrdersShow() {
	as.False(false)
}

func (as *ActionSuite) Test_OrdersCreate() {
	as.False(false)
}

func (as *ActionSuite) Test_OrdersUpdate() {
	as.False(false)
}

func (as *ActionSuite) Test_OrdersDestroy() {
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
	customer := as.getCustomer("UEFA Liv")
	as.Equal(firmino.TenantID, customer.TenantID)

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			var name = fmt.Sprintf("order%s", test.username)
			neworder := &models.Order{SerialNumber: name, Status: "Open", TenantID: firmino.TenantID, CustomerID: customer.ID, CreatedBy: firmino.ID}
			v, err := as.DB.ValidateAndCreate(neworder)
			as.Nil(err)
			as.Equal(0, len(v.Errors))

			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, fmt.Sprintf("/orders/%s", neworder.ID))
			res := req.Delete()
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusOK {
				var order = models.Order{}
				res.Bind(&order)
				as.Equal(name, order.SerialNumber)
				// Check if actually deleted
				order = models.Order{}
				err = as.DB.Where("serial_number=?", name).First(&order)
				as.Equal(err, sql.ErrNoRows)
			} else {
				order := models.Order{}
				err = as.DB.Where("serial_number=?", name).First(&order)
				//Not deleted yet
				as.Equal(name, order.SerialNumber)
			}
		})
	}
}
