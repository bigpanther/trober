package actions

import (
	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/httptest"
	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
)

func (as *ActionSuite) getLoggedInUser(username string) *models.User {
	var user = &models.User{}
	err := as.DB.Where("username=?", username).First(user)
	as.NoError(err)
	as.NotZero(user.ID)
	return user
}

func (as *ActionSuite) getCustomer(name string) *models.Customer {
	var customer = &models.Customer{}
	err := as.DB.Where("name=?", name).First(customer)
	as.NoError(err)
	as.NotZero(customer.ID)
	return customer
}

func (as *ActionSuite) setupRequest(user *models.User, route string) *httptest.JSON {
	req := as.JSON(route)
	req.Headers[xToken] = user.Username
	return req
}

func (as *ActionSuite) createTerminal(name string, terminalType models.TerminalType, tenantID uuid.UUID, createdBy uuid.UUID) *models.Terminal {
	newTerminal := &models.Terminal{Name: name, Type: terminalType.String(), TenantID: tenantID, CreatedBy: createdBy}
	v, err := as.DB.ValidateAndCreate(newTerminal)
	as.Nil(err)
	as.Equal(0, len(v.Errors))
	return newTerminal
}
func (as *ActionSuite) createShipment(shipment models.Shipment, order *models.Order) *models.Shipment {
	if order != nil {
		shipment.OrderID = nulls.NewUUID(order.ID)
		shipment.CustomerID = nulls.NewUUID(order.CustomerID)
	}
	v, err := as.DB.ValidateAndCreate(&shipment)
	as.Nil(err)
	as.Equal(0, len(v.Errors))
	return &shipment
}

func (as *ActionSuite) createCarrier(name string, carrierType models.CarrierType, eta nulls.Time, tenantID uuid.UUID, createdBy uuid.UUID) *models.Carrier {
	newCarrier := &models.Carrier{Name: name, Type: carrierType.String(), TenantID: tenantID, CreatedBy: createdBy, Eta: eta}
	v, err := as.DB.ValidateAndCreate(newCarrier)
	as.Nil(err)
	as.Equal(0, len(v.Errors))
	return newCarrier
}

func (as *ActionSuite) createCustomer(name string, tenantID uuid.UUID, createdBy nulls.UUID) *models.Customer {
	newCustomer := &models.Customer{Name: name, TenantID: tenantID, CreatedBy: createdBy}
	v, err := as.DB.ValidateAndCreate(newCustomer)
	as.Nil(err)
	as.Equal(0, len(v.Errors))
	return newCustomer
}
func (as *ActionSuite) createOrder(serialNumber string, orderStatus models.OrderStatus, tenantID uuid.UUID, createdBy uuid.UUID, customerID uuid.UUID) *models.Order {
	newOrder := &models.Order{SerialNumber: serialNumber, Status: orderStatus.String(), TenantID: tenantID, CreatedBy: createdBy, CustomerID: customerID}
	v, err := as.DB.ValidateAndCreate(newOrder)
	as.Nil(err)
	as.Equal(0, len(v.Errors))
	return newOrder
}

func (as *ActionSuite) createUser(name string, role models.UserRole, email string, tenantID uuid.UUID, customerID nulls.UUID) *models.User {
	newUser := &models.User{Name: name, Role: role.String(), Username: name, Email: email, TenantID: tenantID, CustomerID: customerID}
	v, err := as.DB.ValidateAndCreate(newUser)
	as.Nil(err)
	as.Equal(0, len(v.Errors))
	return newUser
}
func fakeSendNotification(args worker.Args) error { return nil }
