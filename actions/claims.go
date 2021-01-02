package actions

import (
	"context"
	"encoding/base64"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/messaging"
	"github.com/bigpanther/trober/models"
	"google.golang.org/api/option"
)

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

// SetClaims sets the custom claims for the user
func SetClaims(c context.Context, u *models.User) error {
	client, err := firebaseClient()
	if err != nil {
		return err
	}
	claims := map[string]interface{}{
		"bpTenantId":   u.TenantID.String(),
		"bpRole":       u.Role,
		"bpCustomerId": u.CustomerID.UUID.String(),
	}
	if !u.CustomerID.Valid {
		claims["bpCustomerId"] = nil
	}
	err = client.authClient.SetCustomUserClaims(c, u.Username, claims)
	if err != nil {
		return err
	}
	return nil
}

// GetClaims returns the claims for the user
func GetClaims(c context.Context, u *models.User) (map[string]interface{}, error) {
	client, err := firebaseClient()
	if err != nil {
		return nil, err
	}
	ur, err := client.authClient.GetUser(c, u.Username)
	if err != nil {
		return nil, err
	}
	return ur.CustomClaims, nil
}
