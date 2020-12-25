package actions

import (
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"github.com/bigpanther/trober/models"
)

func (as *ActionSuite) Test_OrdersList() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
		orderCount   int
	}{
		{"klopp", http.StatusOK, 1},
		{"firmino", http.StatusOK, 1},
		{"mane", http.StatusOK, 1},
		{"salah", http.StatusNotFound, 1},
		{"nike", http.StatusOK, 1},
		{"coutinho", http.StatusNotFound, 0},
		{"richarlson", http.StatusOK, 0},
		{"rodriguez", http.StatusOK, 0},
		{"lewin", http.StatusNotFound, 0},
		{"allan", http.StatusNotFound, 0},
		{"adidas", http.StatusOK, 0},
	}
	firmino := as.getLoggedInUser("firmino")
	efaLiv := as.getCustomer("EFA Liv")
	as.Equal(efaLiv.TenantID, firmino.TenantID)
	newOrder := as.createOrder("order", models.OrderStatusOpen, firmino.TenantID, firmino.ID, efaLiv.ID)

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, "/orders")
			res := req.Get()
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusOK {
				var orders = models.Orders{}
				res.Bind(&orders)
				as.Equal(test.orderCount, len(orders))
				if test.orderCount > 0 {
					as.Equal(newOrder.SerialNumber, orders[0].SerialNumber)
				}
			}
		})
	}
}

func (as *ActionSuite) Test_OrdersListFilter() {
	as.LoadFixture("Tenant bootstrap")
	var username = "firmino"
	user := as.getLoggedInUser(username)
	efaLiv := as.getCustomer("EFA Liv")
	as.Equal(efaLiv.TenantID, user.TenantID)

	var prefixes = []string{"ਪੰਜਾਬੀ", "Test"}
	for _, p := range prefixes {
		for i := 0; i < 5; i++ {
			orderStatus := models.OrderStatusOpen
			if i%2 == 0 {
				orderStatus = models.OrderStatusAccepted
			}
			_ = as.createOrder(fmt.Sprintf("%s-%d", p, i), orderStatus, user.TenantID, user.ID, efaLiv.ID)
		}
	}
	req := as.setupRequest(user, "/orders?serial_number=ਪੰ&status=Open")
	res := req.Get()
	as.Equal(http.StatusOK, res.Code)
	var orders = models.Orders{}
	res.Bind(&orders)
	as.Equal(2, len(orders))
	for _, v := range orders {
		as.Contains(v.SerialNumber, "ਪੰਜਾਬੀ")
		as.Equal(models.OrderStatusOpen.String(), v.Status)
	}
	klopp := as.getLoggedInUser("klopp")
	as.NotEqual(klopp.TenantID, user.TenantID)

	as.False(user.IsSuperAdmin())
	req = as.setupRequest(user, fmt.Sprintf("/orders?tenant_id=%s", klopp.TenantID))
	res = req.Get()
	as.Equal(http.StatusOK, res.Code)
	orders = models.Orders{}
	res.Bind(&orders)
	as.Equal(0, len(orders))

	lewin := as.getLoggedInUser("lewin")
	as.NotEqual(klopp.TenantID, lewin.TenantID)
	as.NotEqual(lewin.TenantID, user.TenantID)
	as.True(klopp.IsSuperAdmin())
	req = as.setupRequest(klopp, fmt.Sprintf("/orders?tenant_id=%s", lewin.TenantID))
	res = req.Get()
	as.Equal(http.StatusOK, res.Code)
	orders = models.Orders{}
	res.Bind(&orders)
	as.Equal(0, len(orders))
}

func (as *ActionSuite) Test_OrdersShow() {
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
		{"nike", http.StatusOK},
		{"adidas", http.StatusOK},
	}
	richarlson := as.getLoggedInUser("richarlson")
	firmino := as.getLoggedInUser("firmino")
	as.NotEqual(firmino.TenantID, richarlson.TenantID)
	efaLiv := as.getCustomer("EFA Liv")
	as.Equal(efaLiv.TenantID, firmino.TenantID)

	efaEve := as.getCustomer("EFA Eve")
	as.Equal(efaEve.TenantID, richarlson.TenantID)

	var orders = []*models.Order{as.createOrder("term1", models.OrderStatusOpen, firmino.TenantID, firmino.ID, efaLiv.ID),
		as.createOrder("term2", models.OrderStatusAccepted, richarlson.TenantID, richarlson.ID, efaEve.ID)}
	as.NotEqual(orders[0].TenantID, orders[1].TenantID)

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			for _, v := range orders {
				req := as.setupRequest(user, fmt.Sprintf("/orders/%s", v.ID))
				res := req.Get()
				if v.TenantID == user.TenantID || user.IsSuperAdmin() {
					as.Equal(test.responseCode, res.Code)
				} else {
					as.Equal(http.StatusNotFound, res.Code)
				}
				if res.Code == http.StatusOK {
					var order = models.Order{}
					res.Bind(&order)
					as.Equal(v.SerialNumber, order.SerialNumber)
					as.Equal(v.Status, order.Status)
				}
			}
		})
	}
}

func (as *ActionSuite) Test_OrdersCreate() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
	}{
		{"mane", http.StatusCreated},
		{"firmino", http.StatusCreated},
		{"rodriguez", http.StatusBadRequest}, // customer id mismatch
		{"allan", http.StatusNotFound},
		{"salah", http.StatusNotFound},
		{"klopp", http.StatusBadRequest}, // customer id mismatch
		{"adidas", http.StatusCreated},
		{"nike", http.StatusCreated},
	}
	var firmino = as.getLoggedInUser("firmino")
	efaLiv := as.getCustomer("EFA Liv")

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			newOrder := models.Order{SerialNumber: user.Username, Status: models.OrderStatusAccepted.String(), TenantID: firmino.TenantID, CustomerID: efaLiv.ID}
			req := as.setupRequest(user, "/orders")
			res := req.Post(newOrder)
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusCreated {
				var order = models.Order{}
				res.Bind(&order)
				as.Equal(newOrder.SerialNumber, order.SerialNumber)
				as.Equal(models.OrderStatusOpen.String(), order.Status)
				as.Equal(user.TenantID, order.TenantID)
				if user.IsCustomer() {
					as.Equal(user.CustomerID.UUID, order.CustomerID)
				} else {
					as.Equal(efaLiv.ID, order.CustomerID)
				}
			}
		})
	}
}

func (as *ActionSuite) Test_OrdersUpdate() {
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
	efaLiv := as.getCustomer("EFA Liv")

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			newOrder := as.createOrder("order", models.OrderStatusOpen, firmino.TenantID, firmino.ID, efaLiv.ID)
			req := as.setupRequest(user, fmt.Sprintf("/orders/%s", newOrder.ID))
			// Try to update ID and tenant ID. Expect these calls to be excluded at update
			updatedOrder := models.Order{SerialNumber: fmt.Sprintf("not%s", test.username), Status: models.OrderStatusAccepted.String(), ID: user.ID, TenantID: user.ID}
			res := req.Put(updatedOrder)
			as.Equal(test.responseCode, res.Code)
			var dbOrder = *newOrder
			err := as.DB.Reload(&dbOrder)
			as.Nil(err)
			if res.Code == http.StatusOK {
				var order = models.Order{}
				res.Bind(&order)
				as.Equal(updatedOrder.SerialNumber, order.SerialNumber)
				as.Equal(updatedOrder.Status, order.Status)
				as.Equal(newOrder.ID, order.ID)
				as.Equal(dbOrder.SerialNumber, order.SerialNumber)
			} else {
				// Ensure update did not succeed
				as.Equal(dbOrder.SerialNumber, newOrder.SerialNumber)
			}
		})
	}
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
	customer := as.getCustomer("EFA Liv")
	as.Equal(firmino.TenantID, customer.TenantID)

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			var name = fmt.Sprintf("order%s", test.username)
			neworder := &models.Order{SerialNumber: name, Status: models.OrderStatusOpen.String(), TenantID: firmino.TenantID, CustomerID: customer.ID, CreatedBy: firmino.ID}
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
