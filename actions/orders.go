package actions

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
)

// Following naming logic is implemented in Buffalo:
// Model: Singular (Order)
// DB Table: Plural (orders)
// Resource: Plural (Orders)
// Path: Plural (/orders)

// ordersList gets all Orders. This function is mapped to the path
// GET /orders
func ordersList(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)
	tx := c.Value("tx").(*pop.Connection)
	orderSerialNumber := c.Param("serial_number")
	orderStatus := c.Param("status")
	orders := &models.Orders{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())

	if orderSerialNumber != "" {
		if len(orderSerialNumber) < 2 {
			return c.Render(http.StatusOK, r.JSON(orders))
		}
		q = q.Where("serial_number ILIKE ?", fmt.Sprintf("%%%s%%", orderSerialNumber))
	}
	customerID := c.Param("customer_id")
	if loggedInUser.IsCustomer() {
		if !loggedInUser.CustomerID.Valid {
			return c.Error(http.StatusNotFound, errors.New("invalid user"))
		}
		customerID = loggedInUser.CustomerID.UUID.String()
	}
	if customerID != "" {
		q = q.Where("customer_id = ?", customerID)
	}
	if orderStatus != "" {
		q = q.Where("status = ?", orderStatus)
	}
	// Retrieve all orders from the DB
	if err := q.Scope(restrictedScope(c)).Order(orderByCreatedAtDesc).All(orders); err != nil {
		return err
	}

	return c.Render(http.StatusOK, r.JSON(orders))

}

// ordersShow gets the data for one Order. This function is mapped to
// the path GET /orders/{order_id}
func ordersShow(c buffalo.Context) error {

	tx := c.Value("tx").(*pop.Connection)
	var loggedInUser = loggedInUser(c)
	customerID := ""
	if loggedInUser.IsCustomer() {
		if !loggedInUser.CustomerID.Valid {
			return c.Error(http.StatusNotFound, errors.New("invalid user"))
		}
		customerID = loggedInUser.CustomerID.UUID.String()
	}

	var populatedFields = []string{"Customer"}
	q := tx.Eager(populatedFields...).Scope(restrictedScope(c))
	if customerID != "" {
		q = q.Where("customer_id = ?", customerID)
	}

	order := &models.Order{}
	if err := q.Find(order, c.Param("order_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	shipmentsCount, err := shipmentsCount(c, tx, order.ID)
	if err != nil {
		return err
	}
	order.ShipmentCount = shipmentsCount
	return c.Render(http.StatusOK, r.JSON(order))

}

// ordersCreate adds a Order to the DB. This function is mapped to the
// path POST /orders
func ordersCreate(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)

	order := &models.Order{}

	// Bind order to request body
	if err := c.Bind(order); err != nil {
		c.Logger().Errorf("error binding order: %v\n", err)

		return err
	}

	order.Status = models.OrderStatusOpen.String()
	order.TenantID = loggedInUser.TenantID
	order.CreatedBy = loggedInUser.ID
	if order.Type == "" {
		order.Type = models.ShipmentTypeInbound.String()
	}

	tx := c.Value("tx").(*pop.Connection)
	if loggedInUser.IsCustomer() {
		order.CustomerID = loggedInUser.CustomerID.UUID
	} else if err := checkCustomerID(c, tx, loggedInUser, order); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}
	if err := checkCarrierID(c, tx, loggedInUser, order.CarrierID); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}
	for i := range order.Shipments {
		order.Shipments[i].TenantID = order.TenantID
		order.Shipments[i].Type = order.Type
		order.Shipments[i].CreatedBy = loggedInUser.ID
	}
	verrs, err := tx.Eager().ValidateAndCreate(order)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}
	shipmentsCount, err := shipmentsCount(c, tx, order.ID)
	if err != nil {
		return err
	}
	order.ShipmentCount = shipmentsCount
	return c.Render(http.StatusCreated, r.JSON(order))

}

// ordersUpdate changes a Order in the DB. This function is mapped to
// the path PUT /orders/{order_id}
func ordersUpdate(c buffalo.Context) error {

	tx := c.Value("tx").(*pop.Connection)

	order := &models.Order{}
	if err := tx.Scope(restrictedScope(c)).Find(order, c.Param("order_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	newOrder := &models.Order{}
	// Bind Order to request body
	if err := c.Bind(newOrder); err != nil {
		c.Logger().Errorf("error binding order: %v\n", err)

		return err
	}
	if newOrder.SerialNumber != order.SerialNumber || newOrder.Status != order.Status || newOrder.Eta != order.Eta || order.Docco != newOrder.Docco || order.ContainterStatus != newOrder.ContainterStatus || order.CarrierID != newOrder.CarrierID || order.TerminalID != newOrder.TerminalID || order.DropoffCharges != newOrder.DropoffCharges || order.DropoffCost != newOrder.DropoffCost || order.PickupCharges != newOrder.PickupCharges || order.PickupCost != newOrder.PickupCost || order.Rld != newOrder.Rld || order.Shipline != newOrder.Shipline || order.Erd != newOrder.Erd || order.Lfd != newOrder.Lfd || order.SoNumber != newOrder.SoNumber {
		order.UpdatedAt = time.Now().UTC()
		order.Eta = newOrder.Eta
		order.Docco = newOrder.Docco
		order.ContainterStatus = newOrder.ContainterStatus
		order.CarrierID = newOrder.CarrierID
		order.TerminalID = newOrder.TerminalID
		order.DropoffCharges = newOrder.DropoffCharges
		order.DropoffCost = newOrder.DropoffCost
		order.PickupCharges = newOrder.PickupCharges
		order.PickupCost = newOrder.PickupCost
		order.Rld = newOrder.Rld
		order.Shipline = newOrder.Shipline
		order.Erd = newOrder.Erd
		order.Lfd = newOrder.Lfd
		order.SoNumber = newOrder.SoNumber
		order.SerialNumber = newOrder.SerialNumber
		order.Status = newOrder.Status
	} else {
		return c.Render(http.StatusOK, r.JSON(order))
	}
	verrs, err := tx.ValidateAndUpdate(order)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}
	return c.Render(http.StatusOK, r.JSON(order))

}

// ordersDestroy deletes a Order from the DB. This function is mapped
// to the path DELETE /orders/{order_id}
func ordersDestroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	order := &models.Order{}

	if err := tx.Scope(restrictedScope(c)).Find(order, c.Param("order_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(order); err != nil {
		return err
	}
	c.Response().WriteHeader(http.StatusNoContent)
	return nil

}
func checkCustomerID(c buffalo.Context, tx *pop.Connection, loggedInUser *models.User, order *models.Order) error {
	customer := &models.Customer{}
	// User must belong to a customer in the same tenant
	err := tx.Scope(restrictedScope(c)).Where("tenant_id = ?", loggedInUser.TenantID).Find(customer, order.CustomerID)
	if err != nil || order.TenantID != customer.TenantID {
		return errors.New("invalid customer association")
	}
	return nil
}
func shipmentsCount(c buffalo.Context, tx *pop.Connection, orderID uuid.UUID) (int, error) {
	shipmentsCount, err := tx.Scope(restrictedScope(c)).Where("order_id = ?", orderID).Count(&models.Shipments{})
	if err != nil {
		c.Logger().Errorf("error retrieving shipment count for order: %v\n", err)
		return 0, err
	}
	return shipmentsCount, err
}
