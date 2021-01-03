package actions

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/bigpanther/trober/firebase"
	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/httptest"
	"github.com/gobuffalo/nulls"
)

func TestTokenVerify(t *testing.T) {
	t.Skip()
	//os.Setenv("FIREBASE_SA_CRED_FILE", "Path to firebase key")
	// cat filename | base64
	var encodedJSON = ""
	os.Setenv("FIREBASE_SERVICE_ACCOUNT_JSON_ENCODED", encodedJSON)
	ctx := context.Background()
	var tokenToVerify = "..---"
	token, err := firebase.VerifyIDToken(ctx, tokenToVerify)
	if err != nil {
		t.Fatalf("error validating token: %v\n", err)
	}
	user, err := firebase.GetUser(ctx, token.Subject)
	if err != nil {
		t.Fatalf("error getting user: %v\n", err)
	}
	//Print the email always
	if user.Email != "test" {
		t.Errorf("found user %s %s %s", user.Email, user.UID, token.Subject)
	}
	fmt.Println(user.Email)
}

func TestCreateOrUpdateUserOnFirstLoginEmailNotVerified(t *testing.T) {
	app := buffalo.New(buffalo.NewOptions())
	app.GET("/", testCreateOrUpdateUserOnFirstLoginHandler(&auth.UserRecord{EmailVerified: false}, nil))
	ht := httptest.New(app)
	ts := httptest.NewServer(ht)
	defer ts.Close()
	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if http.StatusForbidden != res.StatusCode {
		t.Fatalf("expect %d, got %d", http.StatusForbidden, res.StatusCode)
	}
}
func (as *ActionSuite) Test_CreateUserOnFirstLogin() {
	as.LoadFixture("Tenant bootstrap")
	var (
		uid   = "username"
		name  = "name"
		email = "testcreate@bigpanther.ca"
	)
	message := make(chan string, 2)
	defer close(message)
	callback := func(adminUser *models.User, newUser *models.User, msg string) {
		message <- msg
	}
	app := as.App
	h := testCreateOrUpdateUserOnFirstLoginHandler(&auth.UserRecord{EmailVerified: true, UserInfo: &auth.UserInfo{
		UID:         uid,
		Email:       email,
		DisplayName: name,
	}}, callback)
	app.Middleware.Skip(setCurrentUser, h)
	app.Middleware.Skip(requireActiveUser, h)
	app.GET("/testcreate", h)
	req := as.JSON("/testcreate")
	res := req.Get()
	as.Equal(http.StatusOK, res.Code, string(res.Body.Bytes()))
	var user = models.User{}
	res.Bind(&user)
	as.Equal(name, user.Name)

	as.Equal(email, user.Email)
	as.Eventually(func() bool {
		msg := <-message
		return msg == "New user created"
	}, time.Second*3, time.Second)
}

func (as *ActionSuite) Test_UpdateUserOnFirstLogin() {
	as.LoadFixture("Tenant bootstrap")
	var (
		uid   = "username"
		name  = "name"
		email = "testcreate@bigpanther.ca"
	)
	message := make(chan string, 2)
	var firmino = as.getLoggedInUser("firmino")
	as.createUser("placeholder", models.UserRoleBackOffice, email, firmino.TenantID, nulls.UUID{})
	defer close(message)
	callback := func(adminUser *models.User, newUser *models.User, msg string) {
		message <- msg
	}
	app := as.App
	h := testCreateOrUpdateUserOnFirstLoginHandler(&auth.UserRecord{EmailVerified: true, UserInfo: &auth.UserInfo{
		UID:         uid,
		Email:       email,
		DisplayName: name,
	}}, callback)
	app.Middleware.Skip(setCurrentUser, h)
	app.Middleware.Skip(requireActiveUser, h)
	app.GET("/testupdated", h)
	req := as.JSON("/testupdated")
	res := req.Get()
	as.Equal(http.StatusOK, res.Code, string(res.Body.Bytes()))
	var user = models.User{}
	res.Bind(&user)
	as.Equal(name, user.Name)
	as.Equal(firmino.TenantID, user.TenantID)
	as.Equal(models.UserRoleBackOffice.String(), user.Role)

	as.Equal(email, user.Email)
	as.Eventually(func() bool {
		msg := <-message
		return msg == "New user created"
	}, time.Second*3, time.Second)
}

func testCreateOrUpdateUserOnFirstLoginHandler(remoteUser *auth.UserRecord, notificationCallback func(adminUser *models.User, newUser *models.User, msg string)) buffalo.Handler {
	return func(c buffalo.Context) error {
		u, err := createOrUpdateUserOnFirstLogin(c, remoteUser, notificationCallback)
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, r.JSON(u))
	}
}
