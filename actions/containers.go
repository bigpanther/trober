package actions

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
)

// This file is generated by Buffalo. It offers a basic structure for
// adding, editing and deleting a page. If your model is more
// complex or you need more than the basic implementation you need to
// edit this file.

// Following naming logic is implemented in Buffalo:
// Model: Singular (Container)
// DB Table: Plural (containers)
// Resource: Plural (Containers)
// Path: Plural (/containers)
// View Template Folder: Plural (/templates/containers/)

// ContainersResource is the resource for the Container model
type ContainersResource struct {
	buffalo.Resource
}

// List gets all Containers. This function is mapped to the path
// GET /containers
func (v ContainersResource) List(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	containers := &models.Containers{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())
	orderID := c.Param("order_id")
	if orderID != "" {
		q = q.Where("order_id = ?", orderID)
	}
	carrierID := c.Param("carrier_id")
	if carrierID != "" {
		q = q.Where("carrier_id = ?", carrierID)
	}
	driverID := c.Param("driver_id")

	var loggedInUser = loggedInUser(c)
	if loggedInUser.IsDriver() {
		driverID = loggedInUser.ID.String()
	}
	if driverID != "" {
		q = q.Where("driver_id = ?", driverID)
	}

	// Retrieve all Containers from the DB
	if err := q.Scope(restrictedScope(c)).Order(orderByCreatedAtDesc).All(containers); err != nil {
		return err
	}

	return c.Render(200, r.JSON(containers))

}

// Show gets the data for one Container. This function is mapped to
// the path GET /containers/{container_id}
func (v ContainersResource) Show(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Container
	container := &models.Container{}
	var populatedFields = []string{"Order", "Driver", "Terminal", "Carrier"}

	// To find the Container the parameter container_id is used.
	if err := tx.Eager(populatedFields...).Scope(restrictedScope(c)).Find(container, c.Param("container_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	return c.Render(200, r.JSON(container))

}

// Create adds a Container to the DB. This function is mapped to the
// path POST /containers
func (v ContainersResource) Create(c buffalo.Context) error {
	// Allocate an empty Container
	container := &models.Container{}

	// Bind container to the html form elements
	if err := c.Bind(container); err != nil {
		return err
	}

	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}
	var loggedInUser = loggedInUser(c)
	if !loggedInUser.IsSuperAdmin() || container.TenantID == uuid.Nil {
		container.TenantID = loggedInUser.TenantID
	}
	container.CreatedBy = loggedInUser.ID
	container.CreatedAt = time.Now().UTC()
	container.UpdatedAt = time.Now().UTC()
	// Validate the data from the html form
	verrs, err := tx.ValidateAndCreate(container)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}

	return c.Render(http.StatusCreated, r.JSON(container))

}

// Update changes a Container in the DB. This function is mapped to
// the path PUT /containers/{container_id}
func (v ContainersResource) Update(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Container
	container := &models.Container{}

	if err := tx.Scope(restrictedScope(c)).Find(container, c.Param("container_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	// Bind Container to the html form elements
	if err := c.Bind(container); err != nil {
		return err
	}
	container.UpdatedAt = time.Now().UTC()

	verrs, err := tx.ValidateAndUpdate(container)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}
	if (container.Status.String == "Assigned" && container.DriverID != nulls.UUID{}) {
		u := &models.User{}
		_ = tx.Where("id = ?", container.DriverID.UUID).First(u)
		if u.DeviceID.String != "" {
			app.Worker.Perform(worker.Job{
				Queue:   "default",
				Handler: "sendNotifications",
				Args: worker.Args{
					"to":            []string{u.DeviceID.String},
					"message.title": fmt.Sprintf("You have been assigned a pickup - %s", container.SerialNumber.String),
					"message.body":  container.SerialNumber.String,
					"message.data": map[string]string{
						"container.id":           container.ID.String(),
						"container.serialNumber": container.SerialNumber.String,
					},
				},
			})
		}
	}

	return c.Render(http.StatusOK, r.JSON(container))

}

// UpdateStatus of a container
func (v ContainersResource) UpdateStatus(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Container
	container := &models.Container{}

	if err := tx.Scope(restrictedScope(c)).Find(container, c.Param("container_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}
	status := c.Param("status")
	if models.IsValidContainerStatus(status) {
		return c.Error(http.StatusBadRequest, errors.New("invalid status"))
	}
	var loggedInUser = loggedInUser(c)
	if loggedInUser.IsAtleastBackOffice() {
		var driverID nulls.UUID

		// Bind driver to the html form elements
		if err := c.Bind(&driverID); err != nil {
			return err
		}
		container.DriverID = driverID
		container.Status.String = status
	}
	if loggedInUser.IsDriver() {
		if container.IsAssigned() {
			container.Status.String = status
		}
		//notify backoffice
	}

	container.UpdatedAt = time.Now().UTC()

	verrs, err := tx.ValidateAndUpdate(container)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
	}

	return c.Render(http.StatusOK, r.JSON(container))

}

// Destroy deletes a Container from the DB. This function is mapped
// to the path DELETE /containers/{container_id}
func (v ContainersResource) Destroy(c buffalo.Context) error {
	var loggedInUser = loggedInUser(c)
	if !loggedInUser.IsAtleastBackOffice() {
		return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(notFound, fmt.Sprint(http.StatusNotFound), errNotFound)))
	}
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return models.ErrNotFound
	}

	// Allocate an empty Container
	container := &models.Container{}

	// To find the Container the parameter container_id is used.
	if err := tx.Scope(restrictedScope(c)).Find(container, c.Param("container_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(container); err != nil {
		return err
	}

	return c.Render(http.StatusOK, r.JSON(container))
}
