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
		{"adidas", http.StatusBadRequest, 0},
	}
	firmino := as.getLoggedInUser("firmino")
	salah := as.getLoggedInUser("salah")
	efaLiv := as.getCustomer("EFA Liv")
	order := as.createOrder("ord1", models.OrderStatusOpen, firmino.TenantID, firmino.ID, efaLiv.ID)
	newShipment := as.createShipment(models.Shipment{SerialNumber: "acb123", Status: models.ShipmentStatusAccepted.String(), Type: models.ShipmentTypeIncoming.String(),
		CreatedBy: firmino.ID, TenantID: firmino.TenantID,
		DriverID: nulls.NewUUID(salah.ID)}, order)

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			var orderIDQuery = ""
			if user.IsCustomer() {
				orderIDQuery = fmt.Sprintf("?order_id=%s", order.ID)
			}
			req := as.setupRequest(user, fmt.Sprintf("/shipments%s", orderIDQuery))
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

func (as *ActionSuite) Test_ShipmentsListFilter() {
	as.LoadFixture("Tenant bootstrap")
	firmino := as.getLoggedInUser("firmino")
	efaLiv := as.getCustomer("EFA Liv")
	as.Equal(efaLiv.TenantID, firmino.TenantID)
	var order = as.createOrder("ord1", models.OrderStatusOpen, firmino.TenantID, firmino.ID, efaLiv.ID)

	var prefixes = []string{"ਪੰਜਾਬੀ", "Test"}
	for _, p := range prefixes {
		for i := 0; i < 5; i++ {
			shipmentStatus := models.ShipmentStatusDelivered
			if i%2 == 0 {
				shipmentStatus = models.ShipmentStatusAccepted
			}
			s := models.Shipment{
				SerialNumber: fmt.Sprintf("%s-%d", p, i), Type: models.ShipmentTypeIncoming.String(), Status: shipmentStatus.String(),
				CreatedBy: firmino.ID, TenantID: firmino.TenantID,
			}
			_ = as.createShipment(s, order)
		}
	}
	nike := as.getLoggedInUser("nike")
	req := as.setupRequest(nike, fmt.Sprintf("/shipments?serial_number=ਪੰ&status=Accepted&order_id=%s", order.ID))
	res := req.Get()
	as.Equal(http.StatusOK, res.Code)
	var shipments = models.Shipments{}
	res.Bind(&shipments)
	as.Equal(3, len(shipments))
	for _, v := range shipments {
		as.Contains(v.SerialNumber, "ਪੰਜਾਬੀ")
		as.Equal(models.ShipmentStatusAccepted.String(), v.Status)
	}
	klopp := as.getLoggedInUser("klopp")
	as.NotEqual(klopp.TenantID, firmino.TenantID)

	as.False(firmino.IsSuperAdmin())
	req = as.setupRequest(firmino, fmt.Sprintf("/shipments?tenant_id=%s", klopp.TenantID))
	res = req.Get()
	as.Equal(http.StatusOK, res.Code)
	shipments = models.Shipments{}
	res.Bind(&shipments)
	as.Equal(0, len(shipments))

	lewin := as.getLoggedInUser("lewin")
	as.NotEqual(klopp.TenantID, lewin.TenantID)
	as.NotEqual(lewin.TenantID, firmino.TenantID)
	as.True(klopp.IsSuperAdmin())
	req = as.setupRequest(klopp, fmt.Sprintf("/shipments?tenant_id=%s", lewin.TenantID))
	res = req.Get()
	as.Equal(http.StatusOK, res.Code)
	shipments = models.Shipments{}
	res.Bind(&shipments)
	as.Equal(0, len(shipments))
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
		as.createShipment(models.Shipment{SerialNumber: "s1", Status: models.ShipmentStatusUnassigned.String(), CreatedBy: firmino.ID, TenantID: firmino.TenantID, Type: models.ShipmentTypeIncoming.String(), DriverID: nulls.NewUUID(salah.ID)}, orders[0]),
		as.createShipment(models.Shipment{SerialNumber: "s2", Status: models.ShipmentStatusUnassigned.String(), CreatedBy: richarlson.ID, TenantID: richarlson.TenantID, Type: models.ShipmentTypeIncoming.String()}, orders[1]),
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
	as.LoadFixture("Tenant bootstrap")
	var tests = []struct {
		username     string
		responseCode int
	}{
		{"mane", http.StatusCreated},
		{"firmino", http.StatusCreated},
		{"rodriguez", http.StatusBadRequest},
		{"allan", http.StatusNotFound},
		{"salah", http.StatusCreated},
		{"klopp", http.StatusCreated},
		{"adidas", http.StatusNotFound},
		{"nike", http.StatusNotFound},
	}
	var firmino = as.getLoggedInUser("firmino")
	efaLiv := as.getCustomer("EFA Liv")
	order := as.createOrder("ord1", models.OrderStatusOpen, firmino.TenantID, firmino.ID, efaLiv.ID)
	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			newShipment := models.Shipment{SerialNumber: user.Username, Type: models.ShipmentTypeIncoming.String(), Status: models.ShipmentStatusDelivered.String(), TenantID: firmino.TenantID, OrderID: nulls.NewUUID(order.ID)}
			req := as.setupRequest(user, "/shipments")
			res := req.Post(newShipment)
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusCreated {
				var shipment = models.Shipment{}
				res.Bind(&shipment)
				as.Equal(newShipment.SerialNumber, shipment.SerialNumber)
				as.Equal(models.ShipmentStatusUnassigned.String(), shipment.Status)
				as.Equal(models.ShipmentTypeIncoming.String(), shipment.Type)
				as.Equal(user.TenantID, shipment.TenantID)
				as.Equal(order.ID, shipment.OrderID.UUID)
			}
		})
	}
}

func (as *ActionSuite) Test_ShipmentsUpdate() {
	as.LoadFixture("Tenant bootstrap")
	as.App.Worker.Register("sendNotifications", fakeSendNotification)
	var tests = []struct {
		username     string
		responseCode int
	}{
		{"mane", http.StatusOK},
		{"firmino", http.StatusOK},
		{"rodriguez", http.StatusNotFound},
		{"coutinho", http.StatusNotFound},
		{"allan", http.StatusNotFound},
		{"salah", http.StatusOK},
		{"klopp", http.StatusOK},
		{"adidas", http.StatusNotFound},
		{"nike", http.StatusNotFound},
	}
	var firmino = as.getLoggedInUser("firmino")
	efaLiv := as.getCustomer("EFA Liv")
	order := as.createOrder("order", models.OrderStatusOpen, firmino.TenantID, firmino.ID, efaLiv.ID)
	salah := as.getLoggedInUser("salah")

	for _, test := range tests {
		as.T().Run(test.username, func(t *testing.T) {
			user := as.getLoggedInUser(test.username)
			newShipment := as.createShipment(models.Shipment{SerialNumber: "s1", Status: models.ShipmentStatusAssigned.String(), CreatedBy: firmino.ID, TenantID: firmino.TenantID, Type: models.ShipmentTypeIncoming.String(), DriverID: nulls.NewUUID(salah.ID)}, order)
			req := as.setupRequest(user, fmt.Sprintf("/shipments/%s", newShipment.ID))
			// Try to update ID and tenant ID. Expect these calls to be excluded at update
			updatedShipment := models.Shipment{SerialNumber: fmt.Sprintf("not%s", test.username), Status: models.ShipmentStatusDelivered.String(), ID: user.ID, TenantID: user.ID}
			res := req.Put(updatedShipment)
			as.Equal(test.responseCode, res.Code)
			var dbShipment = *newShipment
			err := as.DB.Reload(&dbShipment)
			as.Nil(err)
			if res.Code == http.StatusOK {
				var shipment = models.Shipment{}
				res.Bind(&shipment)
				if user.IsAtLeastBackOffice() {
					as.Equal(updatedShipment.SerialNumber, shipment.SerialNumber)
				}
				as.Equal(updatedShipment.Status, shipment.Status)
				as.Equal(newShipment.ID, shipment.ID)
				as.Equal(dbShipment.SerialNumber, shipment.SerialNumber)
			} else {
				// Ensure update did not succeed
				as.Equal(dbShipment.SerialNumber, newShipment.SerialNumber)
				as.Equal(dbShipment.Status, newShipment.Status)
			}
		})
	}
}

func (as *ActionSuite) Test_ShipmentsDestroy() {
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
			var name = fmt.Sprintf("shipment%s", test.username)
			s := models.Shipment{SerialNumber: name, Type: models.ShipmentTypeIncoming.String(), Status: models.ShipmentStatusAccepted.String(),
				CreatedBy: firmino.ID, TenantID: firmino.TenantID}
			newShipment := as.createShipment(s, nil)

			user := as.getLoggedInUser(test.username)
			req := as.setupRequest(user, fmt.Sprintf("/shipments/%s", newShipment.ID))
			res := req.Delete()
			as.Equal(test.responseCode, res.Code)
			if res.Code == http.StatusNoContent {
				// Check if actually deleted
				shipment := models.Shipment{}
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
