package actions

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"firebase.google.com/go/v4/auth"
	messaging "firebase.google.com/go/v4/messaging"
	"github.com/bigpanther/trober/firebase"
	"github.com/bigpanther/trober/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/httptest"
	"github.com/gobuffalo/nulls"
	"github.com/golang/mock/gomock"
)

func TestTokenVerify(t *testing.T) {
	t.Skip()
	//os.Setenv("FIREBASE_SA_CRED_FILE", "Path to firebase key")
	// cat filename | base64
	var encodedJSON = ""
	os.Setenv("FIREBASE_SERVICE_ACCOUNT_JSON_ENCODED", encodedJSON)
	ctx := context.Background()
	var tokenToVerify = "..---"
	f, err := firebase.New()
	if err != nil {
		t.Fatalf("error connecting to firebase: %v\n", err)
	}
	user, err := f.GetUser(ctx, tokenToVerify)
	if err != nil {
		t.Fatalf("error getting user: %v\n", err)
	}
	//Print the email always
	if user.Email != "test" {
		t.Errorf("found user %s %s", user.Email, user.UID)
	}
	fmt.Println(user.Email)
}

func Test_createOrUpdateUserOnFirstLoginInvalidEmail(t *testing.T) {
	app := buffalo.New(buffalo.NewOptions())
	app.GET("/", testCreateOrUpdateUserOnFirstLoginHandler(&auth.UserRecord{UserInfo: &auth.UserInfo{Email: "doesnotexist@bigpanther.ca"}}))
	ht := httptest.New(app)
	ts := httptest.NewServer(ht)
	defer ts.Close()
	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if http.StatusInternalServerError != res.StatusCode {
		t.Fatalf("expect %d, got %d", http.StatusInternalServerError, res.StatusCode)
	}
}

func Test_getCurrentUserFromToken(t *testing.T) {
	app := buffalo.New(buffalo.NewOptions())
	app.GET("/", testGetCurrentUserFromToken(mockFirebase))
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
	var tests = []string{"create", "update"}
	var firmino = as.getLoggedInUser("firmino")
	messageChan := make(chan string, 2)
	defer close(messageChan)
	for _, test := range tests {
		as.T().Run(test, func(t *testing.T) {
			var (
				uid   = test
				name  = "name"
				email = fmt.Sprintf("test%s@bigpanther.ca", test)
			)

			if test == "update" {
				as.createUser("placeholder", models.UserRoleBackOffice, email, firmino.TenantID, nulls.UUID{})
			}
			mockFirebase.EXPECT().SendAll(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
				func(c context.Context, messages []*messaging.Message) error {
					messageChan <- messages[0].Notification.Title
					return nil
				},
			)
			h := testCreateOrUpdateUserOnFirstLoginHandler(&auth.UserRecord{UserInfo: &auth.UserInfo{
				UID:         uid,
				Email:       email,
				DisplayName: name,
			}})
			app := as.App
			app.Middleware.Skip(setCurrentUser(mockFirebase), h)
			app.Middleware.Skip(requireActiveUser, h)

			app.GET("/test"+test, h)
			req := as.JSON("/test" + test)
			res := req.Get()
			as.Equal(http.StatusOK, res.Code, res.Body.String())
			var user = models.User{}
			res.Bind(&user)
			as.Equal(name, user.Name)

			if test == "update" {
				as.Equal(models.UserRoleBackOffice.String(), user.Role)
				as.Equal(firmino.TenantID, user.TenantID)
			} else {
				as.Equal(models.UserRoleNone.String(), user.Role)
			}
			as.Equal(email, user.Email)
			as.Eventually(func() bool {
				msg := <-messageChan
				return msg == "New user created"
			}, time.Second*3, time.Second)
		})
	}
}

// func (as *ActionSuite) Test_UpdateUserOnFirstLogin() {
// 	as.LoadFixture("Tenant bootstrap")
// 	var (
// 		uid   = "username"
// 		name  = "name"
// 		email = "testcreate@bigpanther.ca"
// 	)
// 	message := make(chan string, 2)
// 	var firmino = as.getLoggedInUser("firmino")
// 	as.createUser("placeholder", models.UserRoleBackOffice, email, firmino.TenantID, nulls.UUID{})
// 	defer close(message)
// 	mockFirebase.EXPECT().SendAll(gomock.Any(), gomock.Any()).Times(1).DoAndReturn(
// 		func(c context.Context, messages []*messaging.Message) error {
// 			message <- messages[0].Notification.Title
// 			return nil
// 		})
// 	app := as.App
// 	h := testCreateOrUpdateUserOnFirstLoginHandler(&auth.UserRecord{EmailVerified: true, UserInfo: &auth.UserInfo{
// 		UID:         uid,
// 		Email:       email,
// 		DisplayName: name,
// 	}})
// 	app.Middleware.Skip(setCurrentUser(mockFirebase), h)
// 	app.Middleware.Skip(requireActiveUser, h)
// 	app.GET("/testupdated", h)
// 	req := as.JSON("/testupdated")
// 	res := req.Get()
// 	as.Equal(http.StatusOK, res.Code, res.Body.String())
// 	var user = models.User{}
// 	res.Bind(&user)
// 	as.Equal(name, user.Name)
// 	as.Equal(firmino.TenantID, user.TenantID)
// 	as.Equal(models.UserRoleBackOffice.String(), user.Role)

// 	as.Equal(email, user.Email)
// 	as.Eventually(func() bool {
// 		msg := <-message
// 		return msg == "New user created"
// 	}, time.Second*3, time.Second)
// }

func testCreateOrUpdateUserOnFirstLoginHandler(remoteUser *auth.UserRecord) buffalo.Handler {
	return func(c buffalo.Context) error {
		u, err := createOrUpdateUserOnFirstLogin(c, remoteUser)
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, r.JSON(u))
	}
}

func testGetCurrentUserFromToken(f firebase.Firebase) buffalo.Handler {
	return func(c buffalo.Context) error {
		u, err := getCurrentUserFromToken(c, f)
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, r.JSON(u))
	}
}
