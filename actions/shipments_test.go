package actions

import (
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/nulls"
)

func (as *ActionSuite) Test_ShipmentsList() {
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username      string
		responseCode  int
		shipmentCount int
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
	salah := as.getLoggedInUser("salah")
	efaLiv := as.getCustomer("EFA Liv")
	order := as.createOrder("ord1", models.OrderStatusOpen, firmino.TenantID, firmino.ID, efaLiv.ID)
	newShipment := as.createShipment(models.Shipment{SerialNumber: "acb123", Status: models.ShipmentStatusAccepted.String(), Type: models.ShipmentTypeIncoming.String(),
		CreatedBy: firmino.ID, TenantID: firmino.TenantID,
		OrderID:  nulls.NewUUID(order.ID),
		DriverID: nulls.NewUUID(salah.ID)})

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, fmt.Sprintf("/shipments?order_id=%s", order.ID))
			res := req.Get()
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusOK {
				var shipments = models.Shipments{}
				res.Bind(&shipments)
				as.Equal(test.shipmentCount, len(shipments))
				if test.shipmentCount > 0 {
					as.Equal(newShipment.SerialNumber, shipments[0].SerialNumber)
				}
			}
		})
	}
}

func (as *ActionSuite) Test_ShipmentsShow() {
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
		{"nike", http.StatusOK},
		{"adidas", http.StatusOK},
	}
	richarlson := as.getLoggedInUser("richarlson")
	firmino := as.getLoggedInUser("firmino")
	salah := as.getLoggedInUser("salah")
	as.NotEqual(firmino.TenantID, richarlson.TenantID)
	efaLiv := as.getCustomer("EFA Liv")
	as.Equal(efaLiv.TenantID, firmino.TenantID)

	efaEve := as.getCustomer("EFA Eve")
	as.Equal(efaEve.TenantID, richarlson.TenantID)

	var orders = []*models.Order{as.createOrder("ord1", models.OrderStatusOpen, firmino.TenantID, firmino.ID, efaLiv.ID),
		as.createOrder("ord2", models.OrderStatusAccepted, richarlson.TenantID, richarlson.ID, efaEve.ID)}
	as.NotEqual(orders[0].TenantID, orders[1].TenantID)
	var shipments = []*models.Shipment{
		as.createShipment(models.Shipment{SerialNumber: "s1", Status: models.ShipmentStatusUnassigned.String(), CreatedBy: firmino.ID, TenantID: firmino.TenantID, Type: models.ShipmentTypeIncoming.String(), OrderID: nulls.NewUUID(orders[0].ID), DriverID: nulls.NewUUID(salah.ID)}),
		as.createShipment(models.Shipment{SerialNumber: "s2", Status: models.ShipmentStatusUnassigned.String(), CreatedBy: richarlson.ID, TenantID: richarlson.TenantID, Type: models.ShipmentTypeIncoming.String(), OrderID: nulls.NewUUID(orders[1].ID)}),
	}
	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			for _, v := range shipments {
				req := as.setupRequest(user, fmt.Sprintf("/shipments/%s", v.ID))
				res := req.Get()
				if v.TenantID == user.TenantID || user.IsSuperAdmin() {
					as.Equal(test.responseCode, res.Code)
				} else {
					as.Equal(http.StatusNotFound, res.Code)
				}
				if res.Code == http.StatusOK {
					var shipment = models.Shipment{}
					res.Bind(&shipment)
					as.Equal(v.SerialNumber, shipment.SerialNumber)
					as.Equal(v.Status, shipment.Status)
				}
			}
		})
	}
}

func (as *ActionSuite) Test_ShipmentsCreate() {
	as.False(false)
}

func (as *ActionSuite) Test_ShipmentsUpdate() {
	as.False(false)
}

func (as *ActionSuite) Test_ShipmentsDestroy() {
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
			var name = fmt.Sprintf("shipment%s", test.username)
			s := models.Shipment{SerialNumber: name, Type: models.ShipmentTypeIncoming.String(), Status: models.ShipmentStatusAccepted.String(),
				CreatedBy: firmino.ID, TenantID: firmino.TenantID}
			newShipment := as.createShipment(s)

			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, fmt.Sprintf("/shipments/%s", newShipment.ID))
			res := req.Delete()
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusOK {
				var shipment = models.Shipment{}
				res.Bind(&shipment)
				as.Equal(name, shipment.SerialNumber)
				// Check if actually deleted
				shipment = models.Shipment{}
				err := as.DB.Where("serial_number = ?", name).First(&shipment)
				as.Equal(err, sql.ErrNoRows)
			} else {
				shipment := models.Shipment{}
				err := as.DB.Where("serial_number = ?", name).First(&shipment)
				as.Nil(err)
				//Not deleted yet
				as.Equal(name, shipment.SerialNumber)
			}
		})
	}
}
