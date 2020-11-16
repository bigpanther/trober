package actions

import (
	"context"
	"database/sql"
	"log"
	"os"

	"encoding/base64"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	forcessl "github.com/gobuffalo/mw-forcessl"
	i18n "github.com/gobuffalo/mw-i18n"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/shipanther/trober/models"
	"github.com/unrolled/secure"
	"google.golang.org/api/option"

	"github.com/gobuffalo/buffalo-pop/v2/pop/popmw"
	contenttype "github.com/gobuffalo/mw-contenttype"
	"github.com/gobuffalo/x/sessions"
	"github.com/rs/cors"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App

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
func App() *buffalo.App {
	if app == nil {
		app = buffalo.New(buffalo.Options{
			Env:          ENV,
			SessionStore: sessions.Null{},
			PreWares: []buffalo.PreWare{
				cors.Default().Handler,
			},
			SessionName: "_trober_session",
		})

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
		app.Use(setCurrentUser)

		app.GET("/", HomeHandler)
		app.Middleware.Skip(setCurrentUser, HomeHandler)
		var tenantGroup = app.Group("/tenants")
		tenantGroup.GET("/", TenantsResource{}.List)
		tenantGroup.GET("/{tenant_id}", TenantsResource{}.Show)
		tenantGroup.POST("/", TenantsResource{}.Create)
		tenantGroup.PATCH("/{tenant_id}", TenantsResource{}.Update)
		tenantGroup.DELETE("/{tenant_id}", TenantsResource{}.Destroy)
		var UserGroup = app.Group("/users")
		UserGroup.GET("/", UsersResource{}.List)
		UserGroup.GET("/{user_id}", UsersResource{}.Show)
		UserGroup.POST("/", UsersResource{}.Create)
		UserGroup.PATCH("/{user_id}", UsersResource{}.Update)
		UserGroup.DELETE("/{user_id}", UsersResource{}.Destroy)
		//app.Resource("/tenants", TenantsResource{})
		//app.Resource("/users", UsersResource{})
		var CustomerGroup = app.Group("/customers")
		CustomerGroup.GET("/", CustomersResource{}.List)
		CustomerGroup.GET("/{customer_id}", CustomersResource{}.Show)
		CustomerGroup.POST("/", CustomersResource{}.Create)
		CustomerGroup.PATCH("/{customer_id}", CustomersResource{}.Update)
		CustomerGroup.DELETE("/{customer_id}", CustomersResource{}.Destroy)
		var TerminalGroup = app.Group("/terminals")
		TerminalGroup.GET("/", TerminalsResource{}.List)
		TerminalGroup.GET("/{terminal_id}", TerminalsResource{}.Show)
		TerminalGroup.POST("/", TerminalsResource{}.Create)
		TerminalGroup.PATCH("/{terminal_id}", TerminalsResource{}.Update)
		TerminalGroup.DELETE("/{terminal_id}", TerminalsResource{}.Destroy)
		var YardGroup = app.Group("/yards")
		YardGroup.GET("/", YardsResource{}.List)
		YardGroup.GET("/{yard_id}", YardsResource{}.Show)
		YardGroup.POST("/", YardsResource{}.Create)
		YardGroup.PATCH("/{yard_id}", YardsResource{}.Update)
		YardGroup.DELETE("/{yard_id}", YardsResource{}.Destroy)
		var CarrierGroup = app.Group("/carriers")
		CarrierGroup.GET("/", CarriersResource{}.List)
		CarrierGroup.GET("/{carrier_id}", CarriersResource{}.Show)
		CarrierGroup.POST("/", CarriersResource{}.Create)
		CarrierGroup.PATCH("/{carrier_id}", CarriersResource{}.Update)
		CarrierGroup.DELETE("/{carrier_id}", CarriersResource{}.Destroy)
		var ContainerGroup = app.Group("/containers")
		ContainerGroup.GET("/", ContainersResource{}.List)
		ContainerGroup.GET("/{container_id}", ContainersResource{}.Show)
		ContainerGroup.POST("/", ContainersResource{}.Create)
		ContainerGroup.PATCH("/{container_id}", ContainersResource{}.Update)
		ContainerGroup.DELETE("/{container_id}", ContainersResource{}.Destroy)
		var OrderGroup = app.Group("/orders")
		OrderGroup.GET("/", OrdersResource{}.List)
		OrderGroup.GET("/{order_id}", OrdersResource{}.Show)
		OrderGroup.POST("/", OrdersResource{}.Create)
		OrderGroup.PATCH("/{order_id}", OrdersResource{}.Update)
		OrderGroup.DELETE("/{order_id}", OrdersResource{}.Destroy)

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
		if u.IsSuperAdmin() {
			return q
		}
		return q.Where("tennat_id = ?", u.TenantID)
	}
}

var client *auth.Client

func firebaseClient() (*auth.Client, error) {
	if client != nil {
		return client, nil
	}
	var credsJSONEncoded = os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON_ENCODED")
	credJSON, err := base64.StdEncoding.DecodeString(credsJSONEncoded)
	if err != nil {
		return nil, err
	}
	opt := option.WithCredentialsJSON(credJSON)
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		client = nil
		return nil, err
	}
	client, err = app.Auth(ctx)
	if err != nil {
		client = nil
		return nil, err
	}
	return client, err
}

const currentUserKey = "current_user"

// setCurrentUser attempts to find a user based on the firebase token in the request headers
// If one is found it is set on the context.
func setCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		userID := c.Request().Header.Get("X-TOKEN")
		if userID == "" {
			return c.Render(403, r.JSON(models.NewCustomError("Access denied. Missing credentials", "403", nil)))
		}
		client, err := firebaseClient()
		if err != nil {
			return c.Render(500, r.JSON(models.NewCustomError("error getting downstream client", "500", err)))
		}
		token, err := client.VerifyIDToken(c.Request().Context(), userID)
		if err != nil {
			return c.Render(403, r.JSON(models.NewCustomError("Access denied. Credential validation failed", "403", err)))
		}
		u := &models.User{}
		tx := c.Value("tx").(*pop.Connection)
		err = tx.Where("username = ?", token.Subject).First(u)
		if err != nil && errors.Cause(err) != sql.ErrNoRows {
			return c.Render(403, r.JSON(models.NewCustomError(err.Error(), "403", err)))
		}
		if u.ID == uuid.Nil {
			remoteUser, err := client.GetUser(c.Request().Context(), token.Subject)
			if err != nil {
				return c.Render(403, r.JSON(models.NewCustomError(err.Error(), "403", errors.Wrap(err, "error fetching user details"))))
			}
			u.Name = remoteUser.DisplayName
			u.Role = "None"
			u.Username = remoteUser.UID
			t := &models.Tenant{}
			err = tx.Where("name = ?", "system").Where("type = ?", "System").First(t)
			if err != nil {
				return c.Render(403, r.JSON(models.NewCustomError(err.Error(), "403", errors.Wrap(err, "Failed to find user tenant"))))
			}
			u.TenantID = t.ID
			err = tx.Save(u)
			if err != nil {
				log.Printf("error creating user on login: %v\n", err)
				return c.Render(403, r.JSON(models.NewCustomError(err.Error(), "403", err)))
			}
		}
		c.Set(currentUserKey, u)

		return next(c)
	}
}

func loggedInUser(c buffalo.Context) *models.User {
	return c.Value(currentUserKey).(*models.User)
}
