package actions

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bigpanther/trober/firebase"
	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
)

// Following naming logic is implemented in Buffalo:
// Model: Singular (Shipment)
// DB Table: Plural (shipments)
// Resource: Plural (Shipments)
// Path: Plural (/shipments)

// shipmentsList gets all Shipments. This function is mapped to the path
// GET /shipments
func shipmentsList(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)

	tx := c.Value("tx").(*pop.Connection)

	shipmentSerialNumber := strings.Trim(c.Param("serial_number"), " '")
	shipmentType := c.Param("type")
	shipmentSize := c.Param("size")
	shipmentStatus := c.Param("status")

	shipments := &models.Shipments{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())

	if shipmentSerialNumber != "" {
		if len(shipmentSerialNumber) < 2 {
			return c.Render(http.StatusOK, r.JSON(shipments))
		}
		q = q.Where("serial_number ILIKE ?", fmt.Sprintf("%s%%", shipmentSerialNumber))
	}
	if shipmentType != "" {
		q = q.Where("type = ?", shipmentType)
	}
	if shipmentSize != "" {
		q = q.Where("size = ?", shipmentSize)
	}
	if shipmentStatus != "" {
		q = q.Where("status = ?", shipmentStatus)
	}

	orderID := c.Param("order_id")
	if orderID != "" || loggedInUser.IsCustomer() {
		if _, err := checkOrderID(c, tx, loggedInUser, orderID); err != nil {
			return c.Error(http.StatusBadRequest, err)
		}
		q = q.Where("order_id = ?", orderID)
	}
	carrierID := c.Param("carrier_id")
	if carrierID != "" {
		q = q.Where("carrier_id = ?", carrierID)
	}

	driverID := c.Param("driver_id")
	if loggedInUser.IsDriver() {
		driverID = loggedInUser.ID.String()
	}
	if driverID != "" {
		q = q.Where("driver_id = ?", driverID)
	}

	// Retrieve all Shipments from the DB
	if err := q.Scope(restrictedScope(c)).Order(orderByCreatedAtDesc).All(shipments); err != nil {
		return err
	}

	return c.Render(http.StatusOK, r.JSON(shipments))

}

// shipmentsShow gets the data for one Shipment. This function is mapped to
// the path GET /shipments/{shipment_id}
func shipmentsShow(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	shipment := &models.Shipment{}
	var loggedInUser = loggedInUser(c)
	var populatedFields = []string{"Order", "Driver", "Terminal", "Carrier"}
	q := tx.Eager(populatedFields...).Scope(restrictedScope(c))
	if loggedInUser.IsDriver() {
		q = q.Where("driver_id = ?", loggedInUser.ID)
	}
	if loggedInUser.IsCustomer() {
		q = q.Where("customer_id = ?", loggedInUser.CustomerID)
	}
	if err := q.Find(shipment, c.Param("shipment_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	return c.Render(http.StatusOK, r.JSON(shipment))
}

// shipmentsCreate adds a Shipment to the DB. This function is mapped to the
// path POST /shipments
func shipmentsCreate(c buffalo.Context) error {
	shipment := &models.Shipment{}
	// Bind shipment to request body
	if err := c.Bind(shipment); err != nil {
		return err
	}
	tx := c.Value("tx").(*pop.Connection)
	var loggedInUser = loggedInUser(c)
	shipment.Status = models.ShipmentStatusUnassigned.String()
	shipment.TenantID = loggedInUser.TenantID
	shipment.CreatedBy = loggedInUser.ID

	order, err := checkOrderID(c, tx, loggedInUser, shipment.OrderID.UUID.String())
	if err != nil {
		return c.Error(http.StatusBadRequest, err)
	}
	shipment.CustomerID = nulls.NewUUID(order.CustomerID)
	if loggedInUser.IsDriver() {
		shipment.DriverID = nulls.NewUUID(loggedInUser.ID)
	} else if err := checkDriverID(c, tx, loggedInUser, shipment.DriverID); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}
	if err := checkTerminalID(c, tx, loggedInUser, shipment.TerminalID); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}
	if err := checkCarrierID(c, tx, loggedInUser, shipment.CarrierID); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}
	verrs, err := tx.ValidateAndCreate(shipment)
	if err != nil {
		return err
	}
	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}
	return c.Render(http.StatusCreated, r.JSON(shipment))

}

// shipmentsUpdate changes a Shipment in the DB. This function is mapped to
// the path PUT /shipments/{shipment_id}
func shipmentsUpdate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	shipment := &models.Shipment{}
	q := tx.Scope(restrictedScope(c))
	var loggedInUser = loggedInUser(c)
	if loggedInUser.IsDriver() {
		q = q.Where("driver_id = ?", loggedInUser.ID)
	}
	if err := q.Find(shipment, c.Param("shipment_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	newShipment := &models.Shipment{}
	if err := c.Bind(newShipment); err != nil {
		return err
	}
	if loggedInUser.IsDriver() {
		newShipment.DriverID = nulls.NewUUID(loggedInUser.ID)
	}
	switch models.ShipmentStatus(newShipment.Status) {
	case models.ShipmentStatusUnassigned:
		fallthrough
	case models.ShipmentStatusInTransit:
		fallthrough
	case models.ShipmentStatusRejected:
		newShipment.DriverID = nulls.UUID{}
		fallthrough
	case models.ShipmentStatusArrived:
		if !loggedInUser.IsBackOffice() {
			return c.Error(http.StatusBadRequest, errors.New("invalid status"))
		}
	}
	if loggedInUser.IsDriver() {
		//readonly fields
		newShipment.ReservationTime = shipment.ReservationTime
		newShipment.TerminalID = shipment.TerminalID
		newShipment.CarrierID = shipment.CarrierID
		newShipment.OrderID = shipment.OrderID
		newShipment.CustomerID = shipment.CustomerID
		newShipment.DriverID = shipment.DriverID
		newShipment.Lfd = shipment.Lfd
		newShipment.Type = shipment.Type
		newShipment.SerialNumber = shipment.SerialNumber
		newShipment.Origin = shipment.Origin
		newShipment.Destination = shipment.Destination
	}
	shouldNotifyCustomer := shipment.Status != newShipment.Status && newShipment.Status == models.ShipmentStatusDelivered.String()
	var changed bool
	if shipment.OrderID != newShipment.OrderID || newShipment.CustomerID != shipment.CustomerID {
		changed = true
		order, err := checkOrderID(c, tx, loggedInUser, newShipment.OrderID.UUID.String())
		if err != nil {
			return c.Error(http.StatusBadRequest, err)
		}
		newShipment.CustomerID = nulls.NewUUID(order.CustomerID)
	}
	if shipment.DriverID != newShipment.DriverID {
		changed = true
		if err := checkDriverID(c, tx, loggedInUser, newShipment.DriverID); err != nil {
			return c.Error(http.StatusBadRequest, err)
		}
	}
	if shipment.TerminalID != newShipment.TerminalID {
		changed = true
		if err := checkTerminalID(c, tx, loggedInUser, newShipment.TerminalID); err != nil {
			return c.Error(http.StatusBadRequest, err)
		}

	}
	if shipment.CarrierID != newShipment.CarrierID {
		changed = true
		if err := checkCarrierID(c, tx, loggedInUser, newShipment.CarrierID); err != nil {
			return c.Error(http.StatusBadRequest, err)
		}
	}
	if changed || shipment.SerialNumber != newShipment.SerialNumber || shipment.Status != newShipment.Status || shipment.Type != newShipment.Type || shipment.ReservationTime != newShipment.ReservationTime || shipment.Origin != newShipment.Origin || shipment.Destination != newShipment.Destination {
		shipment.UpdatedAt = time.Now().UTC()
		shipment.Status = newShipment.Status
		shipment.DriverID = newShipment.DriverID
		shipment.ReservationTime = newShipment.ReservationTime
		shipment.SerialNumber = newShipment.SerialNumber
		shipment.TerminalID = newShipment.TerminalID
		shipment.CarrierID = newShipment.CarrierID
		shipment.OrderID = newShipment.OrderID
		shipment.CustomerID = newShipment.CustomerID
		shipment.DriverID = newShipment.DriverID
		shipment.Lfd = newShipment.Lfd
		shipment.Type = newShipment.Type
		shipment.SerialNumber = newShipment.SerialNumber
		shipment.Origin = newShipment.Origin
		shipment.Destination = newShipment.Destination
	} else {
		return c.Render(http.StatusOK, r.JSON(shipment))
	}
	verrs, err := tx.ValidateAndUpdate(shipment)
	if err != nil {
		return err
	}
	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}
	if shouldNotifyCustomer {
		if shipment.CustomerID.Valid {
			app.Worker.Perform(worker.Job{
				Queue:   "default",
				Handler: "sendNotifications",
				Args: worker.Args{
					"topics":        []string{firebase.GetCustomerTopic(loggedInUser.TenantID.String(), shipment.CustomerID.UUID.String())},
					"message.title": fmt.Sprintf("Your shipment has been delivered - %s", shipment.SerialNumber),
					"message.body":  shipment.SerialNumber,
					"message.data": map[string]string{
						"shipment.id":           shipment.ID.String(),
						"shipment.serialNumber": shipment.SerialNumber,
					},
				},
			})
		}
	}
	if loggedInUser.IsDriver() {
		app.Worker.Perform(worker.Job{
			Queue:   "default",
			Handler: "sendNotifications",
			Args: worker.Args{
				"topics":        []string{firebase.GetBackOfficeTopic(loggedInUser)},
				"message.title": fmt.Sprintf("Shipment updated by driver - %s: %s", shipment.SerialNumber, shipment.Status),
				"message.body":  shipment.SerialNumber,
				"message.data": map[string]string{
					"shipment.id":           shipment.ID.String(),
					"shipment.serialNumber": shipment.SerialNumber,
					"shipment.status":       shipment.Status,
				},
			},
		})
	}
	if loggedInUser.IsAtLeastBackOffice() {
		if shipment.DriverID.Valid && (shipment.Status != models.ShipmentStatusAssigned.String() || shipment.Status != models.ShipmentStatusAccepted.String()) {
			message := fmt.Sprintf("You have been assigned a pickup - %s", shipment.SerialNumber)
			if shipment.Status != models.ShipmentStatusAccepted.String() {
				message = fmt.Sprintf("Your assignment has been updated - %s", shipment.SerialNumber)
			}
			app.Worker.Perform(worker.Job{
				Queue:   "default",
				Handler: "sendNotifications",
				Args: worker.Args{
					"topics":        []string{firebase.GetDriverTopic(loggedInUser.TenantID.String(), shipment.DriverID.UUID.String())},
					"message.title": message,
					"message.body":  shipment.SerialNumber,
					"message.data": map[string]string{
						"shipment.id":           shipment.ID.String(),
						"shipment.serialNumber": shipment.SerialNumber,
					},
				},
			})
		}
	}
	return c.Render(http.StatusOK, r.JSON(shipment))
}

// shipmentsDestroy deletes a Shipment from the DB. This function is mapped
// to the path DELETE /shipments/{shipment_id}
func shipmentsDestroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	shipment := &models.Shipment{}

	if err := tx.Scope(restrictedScope(c)).Find(shipment, c.Param("shipment_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(shipment); err != nil {
		return err
	}
	c.Response().WriteHeader(http.StatusNoContent)
	return nil
}

func checkOrderID(c buffalo.Context, tx *pop.Connection, loggedInUser *models.User, orderID string) (order *models.Order, err error) {
	q := tx.Scope(restrictedScope(c))
	order = &models.Order{}
	if loggedInUser.IsCustomer() {
		q = q.Where("customer_id = ?", loggedInUser.CustomerID.UUID)
	} else if orderID == uuid.Nil.String() {
		return order, nil
	}
	// User must belong to a customer in the same tenant
	err = q.Find(order, orderID)
	if err != nil || order.ID == uuid.Nil {
		return nil, errors.New("invalid order association")
	}
	return order, nil
}

func checkDriverID(c buffalo.Context, tx *pop.Connection, loggedInUser *models.User, ID nulls.UUID) error {
	if !ID.Valid {
		return nil
	}
	driver := &models.User{}
	// User must belong to the same tenant
	err := tx.Scope(restrictedScope(c)).Find(driver, ID)
	if err != nil || driver.ID == uuid.Nil {
		return errors.New("invalid driver association")
	}
	return nil
}
func checkTerminalID(c buffalo.Context, tx *pop.Connection, loggedInUser *models.User, ID nulls.UUID) error {
	if !ID.Valid {
		return nil
	}
	terminal := &models.Terminal{}
	// Terminal must belong to the same tenant
	err := tx.Scope(restrictedScope(c)).Find(terminal, ID)
	if err != nil || terminal.ID == uuid.Nil {
		return errors.New("invalid terminal association")
	}
	return nil
}
func checkCarrierID(c buffalo.Context, tx *pop.Connection, loggedInUser *models.User, ID nulls.UUID) error {
	if !ID.Valid {
		return nil
	}
	carrier := &models.Carrier{}
	// Carrier must belong to the same tenant
	err := tx.Scope(restrictedScope(c)).Find(carrier, ID)
	if err != nil || carrier.ID == uuid.Nil {
		return errors.New("invalid carrier association")
	}
	return nil
}
