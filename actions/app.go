package actions

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/bigpanther/trober/firebase"
	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	forcessl "github.com/gobuffalo/mw-forcessl"
	i18n "github.com/gobuffalo/mw-i18n"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/unrolled/secure"

	"github.com/gobuffalo/buffalo-pop/v2/pop/popmw"
	contenttype "github.com/gobuffalo/mw-contenttype"
	"github.com/gobuffalo/x/sessions"
	"github.com/rs/cors"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App

const xToken = "X-TOKEN"

// T is the translator
var T *i18n.Translator

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
//
// Routing, middleware, groups, etc... are declared TOP -> DOWN.
// This means if you add a middleware to `app` *after* declaring a
// group, that group will NOT have that new middleware. The same
// is true of resource declarations as well.
//
// It also means that routes are checked in the order they are declared.
// `ServeFiles` is a CATCH-ALL route, so it should always be
// placed last in the route declarations, as it will prevent routes
// declared after it to never be called.
func App(f firebase.Firebase) *buffalo.App {
	if app == nil {
		if f == nil {
			log.Fatalln("firebase.Firebase cannot be nil")
		}
		app = buffalo.New(buffalo.Options{
			Env:          ENV,
			SessionStore: sessions.Null{},
			PreWares: []buffalo.PreWare{
				cors.AllowAll().Handler,
			},
		})
		app.ErrorHandlers[0] = func(status int, err error, c buffalo.Context) error {
			return c.Render(status, r.JSON(models.NewCustomError(err.Error(), fmt.Sprintf("%d", status), err)))
		}
		app.ErrorHandlers[500] = app.ErrorHandlers[0]
		app.ErrorHandlers[404] = app.ErrorHandlers[0]

		// Automatically redirect to SSL
		app.Use(forceSSL())

		// Log request parameters (filters apply).
		app.Use(paramlogger.ParameterLogger)
		// Set the request content type to JSON
		app.Use(contenttype.Set("application/json"))

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.Connection)
		// Remove to disable this.
		app.Use(popmw.Transaction(models.DB))
		app.Use(setCurrentUser(f), requireActiveUser)

		app.GET("/", homeHandler)
		app.GET("/appinfo", appInfoHandler)
		app.Middleware.Skip(setCurrentUser(f), homeHandler, appInfoHandler)

		app.Middleware.Skip(requireActiveUser, homeHandler, appInfoHandler, selfGet, selfGetTenant)
		var selfGroup = app.Group("/self")
		selfGroup.GET("/", selfGet)
		selfGroup.GET("/tenant", selfGetTenant)
		selfGroup.POST("/device-register", selfPostDeviceRegister(f))
		selfGroup.POST("/device-remove", selfPostDeviceRemove(f))
		var tenantGroup = app.Group("/tenants")
		tenantGroup.GET("/", requireSuperAdminUser(tenantsList))
		tenantGroup.GET("/{tenant_id}", requireSuperAdminUser(tenantsShow))
		tenantGroup.POST("/", requireSuperAdminUser(tenantsCreate))
		tenantGroup.PUT("/{tenant_id}", requireSuperAdminUser(tenantsUpdate))
		tenantGroup.DELETE("/{tenant_id}", requireSuperAdminUser(tenantsDestroy))
		var userGroup = app.Group("/users")
		userGroup.GET("/", requireAtLeastBackOfficeUser(usersList))
		userGroup.GET("/{user_id}", requireAtLeastBackOfficeUser(usersShow))
		userGroup.POST("/", requireAtLeastBackOfficeUser(usersCreate))
		userGroup.PUT("/{user_id}", requireAtLeastBackOfficeUser(usersUpdate))
		userGroup.DELETE("/{user_id}", requireAtLeastBackOfficeUser(usersDestroy))
		var customerGroup = app.Group("/customers")
		customerGroup.GET("/", requireAtLeastBackOfficeUser(customersList))
		customerGroup.GET("/{customer_id}", requireAtLeastCustomerUser(customersShow))
		customerGroup.POST("/", requireAtLeastBackOfficeUser(customersCreate))
		customerGroup.PUT("/{customer_id}", requireAtLeastBackOfficeUser(customersUpdate))
		customerGroup.DELETE("/{customer_id}", requireAtLeastBackOfficeUser(customersDestroy))
		var terminalGroup = app.Group("/terminals")
		terminalGroup.GET("/", terminalsList)
		terminalGroup.GET("/{terminal_id}", terminalsShow)
		terminalGroup.POST("/", requireAtLeastBackOfficeUser(terminalsCreate))
		terminalGroup.PUT("/{terminal_id}", requireAtLeastBackOfficeUser(terminalsUpdate))
		terminalGroup.DELETE("/{terminal_id}", requireAtLeastBackOfficeUser(terminalsDestroy))
		var carrierGroup = app.Group("/carriers")
		carrierGroup.GET("/", carriersList)
		carrierGroup.GET("/{carrier_id}", carriersShow)
		carrierGroup.POST("/", requireAtLeastBackOfficeUser(carriersCreate))
		carrierGroup.PUT("/{carrier_id}", requireAtLeastBackOfficeUser(carriersUpdate))
		carrierGroup.DELETE("/{carrier_id}", requireAtLeastBackOfficeUser(carriersDestroy))
		var shipmentGroup = app.Group("/shipments")
		shipmentGroup.GET("/", shipmentsList)
		shipmentGroup.GET("/{shipment_id}", shipmentsShow)
		shipmentGroup.POST("/", requireAtLeastDriverUser(shipmentsCreate))
		shipmentGroup.PUT("/{shipment_id}", requireAtLeastDriverUser(shipmentsUpdate))
		shipmentGroup.DELETE("/{shipment_id}", requireAtLeastBackOfficeUser(shipmentsDestroy))
		var orderGroup = app.Group("/orders")
		orderGroup.GET("/", requireAtLeastCustomerUser(ordersList))
		orderGroup.GET("/{order_id}", requireAtLeastCustomerUser(ordersShow))
		orderGroup.POST("/", requireAtLeastCustomerUser(ordersCreate))
		orderGroup.PUT("/{order_id}", requireAtLeastBackOfficeUser(ordersUpdate))
		orderGroup.DELETE("/{order_id}", requireAtLeastBackOfficeUser(ordersDestroy))

		app.Worker.Register("sendNotifications", sendNotifications(f))
		app.Worker.Register("testWorker", testWorker)
	}

	return app
}

// translations will load locale files, set up the translator `actions.T`,
// and will return a middleware to use to load the correct locale for each
// request.
// for more information: https://gobuffalo.io/en/docs/localization
func translations() buffalo.MiddlewareFunc {
	var err error
	if T, err = i18n.New(packr.New("app:locales", "../locales"), "en-US"); err != nil {
		app.Stop(err)
	}
	return T.Middleware()
}

// forceSSL will return a middleware that will redirect an incoming request
// if it is not HTTPS. "http://example.com" => "https://example.com".
// This middleware does **not** enable SSL. for your application. To do that
// we recommend using a proxy: https://gobuffalo.io/en/docs/proxy
// for more information: https://github.com/unrolled/secure/
func forceSSL() buffalo.MiddlewareFunc {
	return forcessl.Middleware(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}

func restrictedScope(c buffalo.Context) pop.ScopeFunc {
	return func(q *pop.Query) *pop.Query {
		u := loggedInUser(c)
		tenantID := c.Request().URL.Query().Get("tenant_id")
		if tenantID != "" && !u.IsSuperAdmin() {
			return q.Where("? = ?", u.TenantID, tenantID)
		}
		if !u.IsSuperAdmin() {
			return q.Where("tenant_id = ?", u.TenantID)
		}
		if tenantID != "" {
			return q.Where("tenant_id = ?", tenantID)
		}
		return q
	}
}

const currentUserKey = "current_user"

func getCurrentUserFromToken(c buffalo.Context, f firebase.Firebase) (*models.User, error) {
	userID := c.Request().Header.Get(xToken)
	if userID == "" {
		return nil, c.Render(http.StatusForbidden, r.JSON(models.NewCustomError("missing credentials", http.StatusText(http.StatusForbidden), nil)))
	}
	firebaseUser, err := f.GetUser(c.Request().Context(), userID)
	if err != nil {
		return nil, c.Render(http.StatusForbidden, r.JSON(models.NewCustomError("credential validation failed", http.StatusText(http.StatusForbidden), err)))
	}
	if !firebaseUser.EmailVerified || firebaseUser.Disabled {
		return nil, c.Render(http.StatusForbidden, r.JSON(models.NewCustomError("user disabled or email not verified", http.StatusText(http.StatusForbidden), nil)))
	}
	var username = firebaseUser.UID
	u := &models.User{}
	tx := c.Value("tx").(*pop.Connection)
	err = tx.Where("username = ?", username).First(u)
	if err != nil && errors.Cause(err) != sql.ErrNoRows {
		return nil, c.Render(http.StatusInternalServerError, r.JSON(models.NewCustomError(err.Error(), http.StatusText(http.StatusInternalServerError), err)))
	}
	if u.ID == uuid.Nil {
		return createOrUpdateUserOnFirstLogin(c, firebaseUser)
	}
	return u, nil
}
func createOrUpdateUserOnFirstLogin(c buffalo.Context, firebaseUser *auth.UserRecord) (*models.User, error) {
	tx := c.Value("tx").(*pop.Connection)
	u := &models.User{}
	// Try to find by email
	err := tx.Where("email = ?", firebaseUser.Email).First(u)
	if err != nil && errors.Cause(err) != sql.ErrNoRows {
		log.Printf("error fetching user by email: %v\n", err)
		return nil, c.Render(http.StatusInternalServerError, r.JSON(models.NewCustomError(err.Error(), http.StatusText(http.StatusInternalServerError), err)))
	}
	var valErrors *validate.Errors
	var topics = []string{firebase.GetSuperAdminTopic()}
	if u.ID == uuid.Nil {
		u = &models.User{Name: firebaseUser.DisplayName, Role: models.UserRoleNone.String(), Username: firebaseUser.UID, Email: firebaseUser.Email}
		t := &models.Tenant{}
		err = tx.Where("type = ?", models.TenantTypeSystem).First(t)
		if err != nil {
			log.Printf("error fetching system tenant: %v\n", err)
			return nil, c.Render(http.StatusInternalServerError, r.JSON(models.NewCustomError(err.Error(), http.StatusText(http.StatusInternalServerError), errors.Wrap(err, "failed to find user tenant"))))
		}
		u.TenantID = t.ID
		valErrors, err = tx.ValidateAndCreate(u)
		if err != nil {
			log.Printf("error creating user on login: %v\n", err)
			return nil, c.Render(http.StatusForbidden, r.JSON(models.NewCustomError(err.Error(), http.StatusText(http.StatusForbidden), err)))
		}
	} else {
		// inivation scenario
		u.Username = firebaseUser.UID
		u.Name = firebaseUser.DisplayName
		valErrors, err = tx.ValidateAndUpdate(u)
		if err != nil {
			log.Printf("error updating user on login: %v\n", err)
			return nil, c.Render(http.StatusForbidden, r.JSON(models.NewCustomError(err.Error(), http.StatusText(http.StatusForbidden), err)))
		}
		topics = []string{firebase.GetSuperAdminTopic(), firebase.GetAdminTopic(u)}
	}
	var message = "New user created"

	if valErrors.HasAny() {
		log.Printf("validation error on user login: %s\n", valErrors.String())
		message = "New user validation failed"
	}
	sendNotificationsAsync(
		topics,
		message,
		fmt.Sprintf("Name: %s", u.Name),
		map[string]string{
			"name": u.Name,
			"id":   u.ID.String(),
		},
	)
	return u, nil
}

func loggedInUser(c buffalo.Context) *models.User {
	return c.Value(currentUserKey).(*models.User)
}
