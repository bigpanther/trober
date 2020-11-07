package actions

import (
	"context"
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
		app.Use(SetCurrentUser)

		app.GET("/", HomeHandler)
		app.Resource("/tenants", TenantsResource{})
		app.Resource("/users", UsersResource{})
		app.Resource("/customers", CustomersResource{})
		app.Resource("/terminals", TerminalsResource{})
		app.Resource("/yards", YardsResource{})
		app.Resource("/carriers", CarriersResource{})
		app.Resource("/containers", ContainersResource{})
		app.Resource("/orders", OrdersResource{})
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
		if u.Role == "SuperAdmin" {
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
	var credsJsonEncoded = os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON_ENCODED")
	credJson, err := base64.StdEncoding.DecodeString(credsJsonEncoded)
	if err != nil {
		return nil, err
	}
	opt := option.WithCredentialsJSON(credJson)
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

// SetCurrentUser attempts to find a user based on the firebase token in the request headers
// If one is found it is set on the context.
func SetCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		userID := c.Request().Header.Get("X-TOKEN")
		if userID == "" {
			return c.Render(403, r.JSON(models.CustomError{Code: "403", Message: "Access denied. Missing credentials"}))
		}
		client, err := firebaseClient()
		if err != nil {
			log.Fatalf("error getting firebase client: %v\n", err)
		}
		token, err := client.VerifyIDToken(c.Request().Context(), userID)
		if err != nil {
			log.Printf("error validating token: %v\n", err)
			return c.Render(403, r.JSON(models.CustomError{Code: "403", Message: "Access denied. Credential validation failed"}))
		}
		// uid, err := uuid.FromString(token.Subject)
		// if err != nil {
		// 	return c.Render(403, r.JSON(models.CustomError{Code: "403", Message: "Access denied. Invalid credentials"}))
		// }
		u := &models.User{}
		tx := c.Value("tx").(*pop.Connection)
		err = tx.Where("username = ?", token.Subject).First(u)
		if err != nil {
			return c.Render(403, r.JSON(models.CustomError{Code: "403", Message: err.Error()}))
		}
		if u.ID == uuid.Nil {
			remoteUser, err := client.GetUser(c.Request().Context(), token.Subject)
			if err != nil {
				log.Printf("error fetching user details: %v\n", err)
				return c.Render(403, r.JSON(models.CustomError{Code: "403", Message: err.Error()}))
			}
			u.Name = remoteUser.DisplayName
			u.Role = "None"
			u.Username = remoteUser.UID
			err = tx.Save(u)
			if err != nil {
				log.Printf("error creating user on login: %v\n", err)
				return c.Render(403, r.JSON(models.CustomError{Code: "403", Message: err.Error()}))
			}
		}
		c.Set(currentUserKey, u)

		return next(c)
	}
}

func loggedInUser(c buffalo.Context) *models.User {
	return c.Value(currentUserKey).(*models.User)
}
