package grifts

import (
	"time"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/nulls"

	"github.com/markbates/grift/grift"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Seeds a database")
	grift.Add("seed", func(c *grift.Context) error {
		tenant := &models.Tenant{
			Name: "system",
			Type: "System",
			Code: "6mapg",
		}
		err := models.DB.Create(tenant)
		if err != nil {
			return err
		}
		user := &models.User{
			Name:     "Big Panther",
			Username: "oaxWWvwxFOM0odE8tJqqdZEYdxG3",
			TenantID: tenant.ID,
			Role:     "SuperAdmin",
			Email:    "info@bigpanther.ca",
		}
		return models.DB.Create(user)
	})
	grift.Desc("demo_create", "Create demo tenant")
	grift.Add("demo_create", func(c *grift.Context) error {
		return demoCreate()
	})

	grift.Desc("demo_drop", "Drop demo tenant")
	grift.Add("demo_drop", func(c *grift.Context) error {
		return demoDrop()
	})

})

func demoCreate() error {
	tenant := &models.Tenant{
		Name: "Acme Enterprises",
		Type: "Test",
		Code: "7acme",
	}
	err := models.DB.Create(tenant)
	if err != nil {
		return err
	}
	user := &models.User{
		Name:     "Admin Kaur",
		Username: "demoadmin",
		TenantID: tenant.ID,
		Role:     "Admin",
		Email:    "demoadmin@bigpanther.ca",
	}
	err = models.DB.Create(user)
	if err != nil {
		return err
	}
	driver := &models.User{
		Name:     "Driver Singh",
		Username: "demodriver",
		TenantID: tenant.ID,
		Role:     "Driver",
		Email:    "demodriver@bigpanther.ca",
	}
	err = models.DB.Create(driver)
	if err != nil {
		return err
	}
	user = &models.User{
		Name:     "Acme None",
		Username: "demonone",
		TenantID: tenant.ID,
		Role:     "None",
		Email:    "demonone@bigpanther.ca",
	}
	err = models.DB.Create(user)
	if err != nil {
		return err
	}
	user = &models.User{
		Name:     "BackOffice Singh",
		Username: "demobackOffice",
		TenantID: tenant.ID,
		Role:     "BackOffice",
		Email:    "demobackOffice@bigpanther.ca",
	}
	err = models.DB.Create(user)
	if err != nil {
		return err
	}
	customer := &models.Customer{
		Name:      "Shipment Enterprises",
		TenantID:  tenant.ID,
		CreatedBy: nulls.NewUUID(user.ID),
	}
	err = models.DB.Create(customer)
	if err != nil {
		return err
	}
	user = &models.User{
		Name:       "Customer Kaur",
		Username:   "democustomer",
		TenantID:   tenant.ID,
		Role:       "Customer",
		CustomerID: nulls.NewUUID(customer.ID),
		Email:      "democustomer@bigpanther.ca",
	}
	err = models.DB.Create(user)
	if err != nil {
		return err
	}
	terminal := &models.Terminal{
		Name:      "Vancouver Port",
		TenantID:  tenant.ID,
		Type:      "Port",
		CreatedBy: user.ID,
	}
	err = models.DB.Create(terminal)
	if err != nil {
		return err
	}
	order := &models.Order{
		SerialNumber: "ORD00001",
		TenantID:     tenant.ID,
		Status:       "Open",
		CustomerID:   customer.ID,
		CreatedBy:    user.ID,
	}
	err = models.DB.Create(order)
	if err != nil {
		return err
	}
	carrier := &models.Carrier{
		Name:      "Global Shippers",
		TenantID:  tenant.ID,
		Eta:       nulls.NewTime(time.Now().AddDate(0, 0, 1)),
		Type:      "Vessel",
		CreatedBy: user.ID,
	}
	err = models.DB.Create(carrier)
	if err != nil {
		return err
	}
	shipment := &models.Shipment{
		SerialNumber:    "CANV2020127",
		TenantID:        tenant.ID,
		Type:            "Incoming",
		Status:          "Unassigned",
		Origin:          nulls.NewString("Seattle"),
		Destination:     nulls.NewString("Hope"),
		Size:            nulls.NewString("40ST"),
		ReservationTime: nulls.NewTime(time.Now().AddDate(0, 0, 2)),
		CarrierID:       nulls.NewUUID(carrier.ID),
		OrderID:         nulls.NewUUID(order.ID),
		TerminalID:      nulls.NewUUID(terminal.ID),
		CreatedBy:       user.ID,
	}
	err = models.DB.Create(shipment)
	shipment = &models.Shipment{
		SerialNumber:    "CANV2020128",
		TenantID:        tenant.ID,
		Type:            "Incoming",
		Status:          "Assigned",
		Origin:          nulls.NewString("Seattle"),
		Destination:     nulls.NewString("Whistler"),
		Size:            nulls.NewString("20ST"),
		ReservationTime: nulls.NewTime(time.Now().AddDate(0, 0, 2)),
		CreatedBy:       user.ID,
		OrderID:         nulls.NewUUID(order.ID),
		CarrierID:       nulls.NewUUID(carrier.ID),
		TerminalID:      nulls.NewUUID(terminal.ID),
		DriverID:        nulls.NewUUID(driver.ID),
	}
	err = models.DB.Create(shipment)
	if err != nil {
		return err
	}
	return nil
}

func demoDrop() error {
	tenant := &models.Tenant{}
	models.DB.Where("name= ?", "Acme Enterprises").Where("type=?", "Test").First(tenant)

	shipments := &models.Shipments{}
	err := models.DB.Where("tenant_id=?", tenant.ID).All(shipments)
	if err != nil {
		return err
	}
	err = models.DB.Destroy(shipments)
	if err != nil {
		return err
	}
	carriers := &models.Carriers{}
	err = models.DB.Where("tenant_id=?", tenant.ID).All(carriers)
	if err != nil {
		return err
	}
	err = models.DB.Destroy(carriers)
	if err != nil {
		return err
	}
	terminals := &models.Terminals{}
	err = models.DB.Where("tenant_id=?", tenant.ID).All(terminals)
	if err != nil {
		return err
	}
	err = models.DB.Destroy(terminals)
	if err != nil {
		return err
	}
	orders := &models.Orders{}
	err = models.DB.Where("tenant_id=?", tenant.ID).All(orders)
	if err != nil {
		return err
	}
	err = models.DB.Destroy(orders)
	if err != nil {
		return err
	}
	users := &models.Users{}
	err = models.DB.Where("tenant_id= ?", tenant.ID).Where("customer_id IS NOT NULL").All(users)
	if err != nil {
		return err
	}
	err = models.DB.Destroy(users)
	customers := &models.Customers{}
	err = models.DB.Where("tenant_id=?", tenant.ID).All(customers)
	if err != nil {
		return err
	}
	err = models.DB.Destroy(customers)
	if err != nil {
		return err
	}
	users = &models.Users{}
	err = models.DB.Where("tenant_id=?", tenant.ID).All(users)
	if err != nil {
		return err
	}
	err = models.DB.Destroy(users)
	if err != nil {
		return err
	}
	err = models.DB.Destroy(tenant)
	if err != nil {
		return err
	}
	return nil
}
