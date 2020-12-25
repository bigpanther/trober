package actions

import (
	"fmt"
	"net/http"

	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
)

// setCurrentUser attempts to find a user based on the firebase token in the request headers
// If one is found it is set on the context.
func setCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		var user *models.User
		var err error
		if ENV == "production" {
			user, err = getCurrentUserFromToken(c)
		} else {
			user = &models.User{}
			tx := c.Value("tx").(*pop.Connection)
			var username = c.Request().Header.Get(xToken)
			err = tx.Where("username = ?", username).First(user)
			if err != nil {
				return c.Render(http.StatusForbidden, r.JSON(models.NewCustomError(err.Error(), http.StatusText(http.StatusForbidden), err)))
			}
		}
		if err != nil {
			return err
		}
		c.Set(currentUserKey, user)
		return next(c)
	}
}

func requireActiveUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		var loggedInUser = loggedInUser(c)
		if loggedInUser.IsNotActive() {
			return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(http.StatusText(http.StatusNotFound), fmt.Sprint(http.StatusNotFound), errNotFound)))
		}
		return next(c)
	}
}

func requireSuperAdminUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		var loggedInUser = loggedInUser(c)
		if !loggedInUser.IsSuperAdmin() {
			return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(http.StatusText(http.StatusNotFound), fmt.Sprint(http.StatusNotFound), errNotFound)))
		}
		return next(c)
	}
}
func requireAtLeastBackOfficeUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		var loggedInUser = loggedInUser(c)
		if !loggedInUser.IsAtLeastBackOffice() {
			return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(http.StatusText(http.StatusNotFound), fmt.Sprint(http.StatusNotFound), errNotFound)))
		}
		return next(c)
	}
}
func requireAtLeastCustomerUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		var loggedInUser = loggedInUser(c)
		if !loggedInUser.IsAtLeastBackOffice() && !loggedInUser.IsCustomer() {
			return c.Render(http.StatusNotFound, r.JSON(models.NewCustomError(http.StatusText(http.StatusNotFound), fmt.Sprint(http.StatusNotFound), errNotFound)))
		}
		return next(c)
	}
}
