package actions

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"encoding/base64"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/messaging"
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

		app.GET("/", homeHandler)
		app.Middleware.Skip(setCurrentUser, homeHandler)
		var tenantGroup = app.Group("/tenants")
		tenantGroup.GET("/", TenantsResource{}.List)
		tenantGroup.GET("/{tenant_id}", TenantsResource{}.Show)
		tenantGroup.POST("/", TenantsResource{}.Create)
		tenantGroup.PUT("/{tenant_id}", TenantsResource{}.Update)
		tenantGroup.DELETE("/{tenant_id}", TenantsResource{}.Destroy)
		var userGroup = app.Group("/users")
		userGroup.GET("/", UsersResource{}.List)
		userGroup.GET("/{user_id}", UsersResource{}.Show)
		userGroup.POST("/", UsersResource{}.Create)
		userGroup.PUT("/{user_id}", UsersResource{}.Update)
		userGroup.DELETE("/{user_id}", UsersResource{}.Destroy)
		//app.Resource("/tenants", TenantsResource{})
		//app.Resource("/users", UsersResource{})
		var customerGroup = app.Group("/customers")
		customerGroup.GET("/", CustomersResource{}.List)
		customerGroup.GET("/{customer_id}", CustomersResource{}.Show)
		customerGroup.POST("/", CustomersResource{}.Create)
		customerGroup.PUT("/{customer_id}", CustomersResource{}.Update)
		customerGroup.DELETE("/{customer_id}", CustomersResource{}.Destroy)
		var terminalGroup = app.Group("/terminals")
		terminalGroup.GET("/", TerminalsResource{}.List)
		terminalGroup.GET("/{terminal_id}", TerminalsResource{}.Show)
		terminalGroup.POST("/", TerminalsResource{}.Create)
		terminalGroup.PUT("/{terminal_id}", TerminalsResource{}.Update)
		terminalGroup.DELETE("/{terminal_id}", TerminalsResource{}.Destroy)
		var carrierGroup = app.Group("/carriers")
		carrierGroup.GET("/", CarriersResource{}.List)
		carrierGroup.GET("/{carrier_id}", CarriersResource{}.Show)
		carrierGroup.POST("/", CarriersResource{}.Create)
		carrierGroup.PUT("/{carrier_id}", CarriersResource{}.Update)
		carrierGroup.DELETE("/{carrier_id}", CarriersResource{}.Destroy)
		var containerGroup = app.Group("/containers")
		containerGroup.GET("/", ContainersResource{}.List)
		containerGroup.GET("/{container_id}", ContainersResource{}.Show)
		containerGroup.POST("/", ContainersResource{}.Create)
		containerGroup.PUT("/{container_id}", ContainersResource{}.Update)
		containerGroup.DELETE("/{container_id}", ContainersResource{}.Destroy)
		var orderGroup = app.Group("/orders")
		orderGroup.GET("/", OrdersResource{}.List)
		orderGroup.GET("/{order_id}", OrdersResource{}.Show)
		orderGroup.POST("/", OrdersResource{}.Create)
		orderGroup.PUT("/{order_id}", OrdersResource{}.Update)
		orderGroup.DELETE("/{order_id}", OrdersResource{}.Destroy)
		var selfGroup = app.Group("/self")
		selfGroup.GET("/", selfGet)
		selfGroup.GET("/tenant", selfGetTenant)

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

		if !u.IsSuperAdmin() {
			return q.Where("tenant_id = ?", u.TenantID)
		}
		tenantID := c.Request().URL.Query().Get("tenant_id")
		if tenantID != "" {
			return q.Where("tenant_id = ?", tenantID)
		}
		return q
	}
}

type firebaseSdkClient struct {
	authClient      *auth.Client
	messagingClient *messaging.Client
}

var client *firebaseSdkClient

func firebaseClient() (*firebaseSdkClient, error) {
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
	client = &firebaseSdkClient{}
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		client = nil
		return nil, err
	}
	client.authClient, err = app.Auth(ctx)
	if err != nil {
		client = nil
		return nil, err
	}
	client.messagingClient, err = app.Messaging(ctx)
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
		var user *models.User
		var err error
		if ENV == "production" {
			user, err = getCurrentUserFromToken(c)
		} else {
			user = &models.User{}
			tx := c.Value("tx").(*pop.Connection)
			var username = c.Request().Header.Get(xToken)
			fmt.Println(username)
			err = tx.Where("username = ?", username).First(user)
			if err != nil {
				return c.Render(403, r.JSON(models.NewCustomError(err.Error(), "403", err)))
			}
		}
		if err != nil {
			return err
		}
		c.Set(currentUserKey, user)
		return next(c)
	}
}

func getCurrentUserFromToken(c buffalo.Context) (*models.User, error) {
	userID := c.Request().Header.Get(xToken)
	if userID == "" {
		return nil, c.Render(403, r.JSON(models.NewCustomError("Access denied. Missing credentials", "403", nil)))
	}
	client, err := firebaseClient()
	if err != nil {
		return nil, c.Render(500, r.JSON(models.NewCustomError("error getting downstream client", "500", err)))
	}
	token, err := client.authClient.VerifyIDToken(c.Request().Context(), userID)
	if err != nil {
		return nil, c.Render(403, r.JSON(models.NewCustomError("Access denied. Credential validation failed", "403", err)))
	}

	u := &models.User{}
	tx := c.Value("tx").(*pop.Connection)
	err = tx.Where("username = ?", token.Subject).First(u)
	if err != nil && errors.Cause(err) != sql.ErrNoRows {
		return nil, c.Render(403, r.JSON(models.NewCustomError(err.Error(), "403", err)))
	}
	if u.ID == uuid.Nil {
		remoteUser, err := client.authClient.GetUser(c.Request().Context(), token.Subject)
		if err != nil {
			return nil, c.Render(403, r.JSON(models.NewCustomError(err.Error(), "403", errors.Wrap(err, "error fetching user details"))))
		}
		if !remoteUser.EmailVerified {
			return nil, c.Render(403, r.JSON(models.NewCustomError("Access denied. Email not verified", "403", err)))
		}
		u.Name = remoteUser.DisplayName
		u.Role = "None"
		u.Username = remoteUser.UID
		u.Email = remoteUser.Email
		t := &models.Tenant{}
		err = tx.Where("name = ?", "system").Where("type = ?", "System").First(t)
		if err != nil {
			return nil, c.Render(403, r.JSON(models.NewCustomError(err.Error(), "403", errors.Wrap(err, "Failed to find user tenant"))))
		}
		u.TenantID = t.ID
		valErrors, err := tx.ValidateAndCreate(u)
		if err != nil {
			log.Printf("error creating user on login: %v\n", err)
			return nil, c.Render(403, r.JSON(models.NewCustomError(err.Error(), "403", err)))
		}
		adminUser := &models.User{}
		_ = tx.Where("tenant_id = ?", t.ID).First(adminUser)
		if valErrors.HasAny() {
			log.Printf("error creating user on login: %s\n", valErrors.String())
			if adminUser != nil {
				sendMessage(adminUser, u, "New user validation failed")
			}
			return nil, c.Render(403, r.JSON(models.NewCustomError(err.Error(), "403", err)))
		}

		if adminUser != nil {
			sendMessage(adminUser, u, "New user created")
		}
	}
	return u, nil
}

func sendMessage(adminUser *models.User, newUser *models.User, msg string) {
	if adminUser.DeviceID.String != "" {
		message := &messaging.Message{
			Data: map[string]string{
				"email": newUser.Email,
				"name":  newUser.Name,
			},
			Notification: &messaging.Notification{
				Title: msg,
				Body:  fmt.Sprintf("%s just logged in", newUser.Name),
			},
			Token: adminUser.DeviceID.String,
		}
		_, err := client.messagingClient.Send(app.Context, message)
		if err != nil {
			log.Println("error sending message", err)
		}
	}
}

func loggedInUser(c buffalo.Context) *models.User {
	return c.Value(currentUserKey).(*models.User)
}
