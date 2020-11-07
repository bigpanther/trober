package actions

import (

  "fmt"
  "net/http"
  "github.com/gobuffalo/buffalo"
  "github.com/gobuffalo/pop/v5"
  "github.com/gobuffalo/x/responder"
  "github.com/shipanther/trober/models"
)

// This file is generated by Buffalo. It offers a basic structure for
// adding, editing and deleting a page. If your model is more
// complex or you need more than the basic implementation you need to
// edit this file.

// Following naming logic is implemented in Buffalo:
// Model: Singular (Order)
// DB Table: Plural (orders)
// Resource: Plural (Orders)
// Path: Plural (/orders)
// View Template Folder: Plural (/templates/orders/)

// OrdersResource is the resource for the Order model
type OrdersResource struct{
  buffalo.Resource
}

// List gets all Orders. This function is mapped to the path
// GET /orders
func (v OrdersResource) List(c buffalo.Context) error {
  // Get the DB connection from the context
  tx, ok := c.Value("tx").(*pop.Connection)
  if !ok {
    return fmt.Errorf("no transaction found")
  }

  orders := &models.Orders{}

  // Paginate results. Params "page" and "per_page" control pagination.
  // Default values are "page=1" and "per_page=20".
  q := tx.PaginateFromParams(c.Params())

  // Retrieve all Orders from the DB
  if err := q.All(orders); err != nil {
    return err
  }

  return responder.Wants("html", func (c buffalo.Context) error {
    // Add the paginator to the context so it can be used in the template.
    c.Set("pagination", q.Paginator)

    c.Set("orders", orders)
    return c.Render(http.StatusOK, r.HTML("/orders/index.plush.html"))
  }).Wants("json", func (c buffalo.Context) error {
    return c.Render(200, r.JSON(orders))
  }).Wants("xml", func (c buffalo.Context) error {
    return c.Render(200, r.XML(orders))
  }).Respond(c)
}

// Show gets the data for one Order. This function is mapped to
// the path GET /orders/{order_id}
func (v OrdersResource) Show(c buffalo.Context) error {
  // Get the DB connection from the context
  tx, ok := c.Value("tx").(*pop.Connection)
  if !ok {
    return fmt.Errorf("no transaction found")
  }

  // Allocate an empty Order
  order := &models.Order{}

  // To find the Order the parameter order_id is used.
  if err := tx.Find(order, c.Param("order_id")); err != nil {
    return c.Error(http.StatusNotFound, err)
  }

  return responder.Wants("html", func (c buffalo.Context) error {
    c.Set("order", order)

    return c.Render(http.StatusOK, r.HTML("/orders/show.plush.html"))
  }).Wants("json", func (c buffalo.Context) error {
    return c.Render(200, r.JSON(order))
  }).Wants("xml", func (c buffalo.Context) error {
    return c.Render(200, r.XML(order))
  }).Respond(c)
}

// Create adds a Order to the DB. This function is mapped to the
// path POST /orders
func (v OrdersResource) Create(c buffalo.Context) error {
  // Allocate an empty Order
  order := &models.Order{}

  // Bind order to the html form elements
  if err := c.Bind(order); err != nil {
    return err
  }

  // Get the DB connection from the context
  tx, ok := c.Value("tx").(*pop.Connection)
  if !ok {
    return fmt.Errorf("no transaction found")
  }

  // Validate the data from the html form
  verrs, err := tx.ValidateAndCreate(order)
  if err != nil {
    return err
  }

  if verrs.HasAny() {
    return responder.Wants("html", func (c buffalo.Context) error {
      // Make the errors available inside the html template
      c.Set("errors", verrs)

      // Render again the new.html template that the user can
      // correct the input.
      c.Set("order", order)

      return c.Render(http.StatusUnprocessableEntity, r.HTML("/orders/new.plush.html"))
    }).Wants("json", func (c buffalo.Context) error {
      return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
    }).Wants("xml", func (c buffalo.Context) error {
      return c.Render(http.StatusUnprocessableEntity, r.XML(verrs))
    }).Respond(c)
  }

  return responder.Wants("html", func (c buffalo.Context) error {
    // If there are no errors set a success message
    c.Flash().Add("success", T.Translate(c, "order.created.success"))

    // and redirect to the show page
    return c.Redirect(http.StatusSeeOther, "/orders/%v", order.ID)
  }).Wants("json", func (c buffalo.Context) error {
    return c.Render(http.StatusCreated, r.JSON(order))
  }).Wants("xml", func (c buffalo.Context) error {
    return c.Render(http.StatusCreated, r.XML(order))
  }).Respond(c)
}

// Update changes a Order in the DB. This function is mapped to
// the path PUT /orders/{order_id}
func (v OrdersResource) Update(c buffalo.Context) error {
  // Get the DB connection from the context
  tx, ok := c.Value("tx").(*pop.Connection)
  if !ok {
    return fmt.Errorf("no transaction found")
  }

  // Allocate an empty Order
  order := &models.Order{}

  if err := tx.Find(order, c.Param("order_id")); err != nil {
    return c.Error(http.StatusNotFound, err)
  }

  // Bind Order to the html form elements
  if err := c.Bind(order); err != nil {
    return err
  }

  verrs, err := tx.ValidateAndUpdate(order)
  if err != nil {
    return err
  }

  if verrs.HasAny() {
    return responder.Wants("html", func (c buffalo.Context) error {
      // Make the errors available inside the html template
      c.Set("errors", verrs)

      // Render again the edit.html template that the user can
      // correct the input.
      c.Set("order", order)

      return c.Render(http.StatusUnprocessableEntity, r.HTML("/orders/edit.plush.html"))
    }).Wants("json", func (c buffalo.Context) error {
      return c.Render(http.StatusUnprocessableEntity, r.JSON(verrs))
    }).Wants("xml", func (c buffalo.Context) error {
      return c.Render(http.StatusUnprocessableEntity, r.XML(verrs))
    }).Respond(c)
  }

  return responder.Wants("html", func (c buffalo.Context) error {
    // If there are no errors set a success message
    c.Flash().Add("success", T.Translate(c, "order.updated.success"))

    // and redirect to the show page
    return c.Redirect(http.StatusSeeOther, "/orders/%v", order.ID)
  }).Wants("json", func (c buffalo.Context) error {
    return c.Render(http.StatusOK, r.JSON(order))
  }).Wants("xml", func (c buffalo.Context) error {
    return c.Render(http.StatusOK, r.XML(order))
  }).Respond(c)
}

// Destroy deletes a Order from the DB. This function is mapped
// to the path DELETE /orders/{order_id}
func (v OrdersResource) Destroy(c buffalo.Context) error {
  // Get the DB connection from the context
  tx, ok := c.Value("tx").(*pop.Connection)
  if !ok {
    return fmt.Errorf("no transaction found")
  }

  // Allocate an empty Order
  order := &models.Order{}

  // To find the Order the parameter order_id is used.
  if err := tx.Find(order, c.Param("order_id")); err != nil {
    return c.Error(http.StatusNotFound, err)
  }

  if err := tx.Destroy(order); err != nil {
    return err
  }

  return responder.Wants("html", func (c buffalo.Context) error {
    // If there are no errors set a flash message
    c.Flash().Add("success", T.Translate(c, "order.destroyed.success"))

    // Redirect to the index page
    return c.Redirect(http.StatusSeeOther, "/orders")
  }).Wants("json", func (c buffalo.Context) error {
    return c.Render(http.StatusOK, r.JSON(order))
  }).Wants("xml", func (c buffalo.Context) error {
    return c.Render(http.StatusOK, r.XML(order))
  }).Respond(c)
}
