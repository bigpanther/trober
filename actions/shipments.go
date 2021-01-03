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
		if err := checkOrderID(c, tx, loggedInUser, orderID); err != nil {
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
	if err := q.Find(shipment, c.Param("shipment_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	if loggedInUser.IsCustomer() && shipment.Order.CustomerID != loggedInUser.CustomerID.UUID {
		return c.Error(http.StatusNotFound, errNotFound)
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
	if err := checkOrderID(c, tx, loggedInUser, shipment.OrderID.UUID.String()); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}
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

	if err := tx.Scope(restrictedScope(c)).Find(shipment, c.Param("shipment_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	// Bind Shipment to request body
	if err := c.Bind(shipment); err != nil {
		return err
	}
	shipment.UpdatedAt = time.Now().UTC()

	verrs, err := tx.ValidateAndUpdate(shipment)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}
	if shipment.Status == "Assigned" && shipment.DriverID.Valid {
		u := &models.User{}
		_ = tx.Where("id = ?", shipment.DriverID.UUID).First(u)
		app.Worker.Perform(worker.Job{
			Queue:   "default",
			Handler: "sendNotifications",
			Args: worker.Args{
				"topics":        []string{firebase.GetTopic(u)},
				"message.title": fmt.Sprintf("You have been assigned a pickup - %s", shipment.SerialNumber),
				"message.body":  shipment.SerialNumber,
				"message.data": map[string]string{
					"shipment.id":           shipment.ID.String(),
					"shipment.serialNumber": shipment.SerialNumber,
				},
			},
		})
	}

	return c.Render(http.StatusOK, r.JSON(shipment))

}

// shipmentsUpdateStatus of a shipment
func shipmentsUpdateStatus(c buffalo.Context) error {

	tx := c.Value("tx").(*pop.Connection)

	shipment := &models.Shipment{}

	if err := tx.Scope(restrictedScope(c)).Find(shipment, c.Param("shipment_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	status := c.Param("status")
	if models.IsValidShipmentStatus(status) {
		return c.Error(http.StatusBadRequest, errors.New("invalid status"))
	}
	var loggedInUser = loggedInUser(c)
	if loggedInUser.IsAtLeastBackOffice() {
		var driverID nulls.UUID

		// Bind driver to request body
		if err := c.Bind(&driverID); err != nil {
			return err
		}
		shipment.DriverID = driverID
		shipment.Status = status
	}
	if loggedInUser.IsDriver() {
		if shipment.IsAssigned() {
			shipment.Status = status
		}
		//notify backoffice
	}

	shipment.UpdatedAt = time.Now().UTC()

	verrs, err := tx.ValidateAndUpdate(shipment)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
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

func checkOrderID(c buffalo.Context, tx *pop.Connection, loggedInUser *models.User, orderID string) error {
	q := tx.Scope(restrictedScope(c))
	if loggedInUser.IsCustomer() {
		q = q.Where("customer_id = ?", loggedInUser.CustomerID.UUID)
	} else if orderID == uuid.Nil.String() {
		return nil
	}
	order := &models.Order{}
	// User must belong to a customer in the same tenant
	err := q.Find(order, orderID)
	if err != nil || order.ID == uuid.Nil {
		return errors.New("invalid order association")
	}
	return nil
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
