package firebase

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
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
		claims["bpCustomerId"] = "none"
	}
	err = client.authClient.SetCustomUserClaims(c, u.Username, claims)
	if err != nil {
		return err
	}
	return nil
}

// VerifyIDToken return the auth token after verification
func VerifyIDToken(c context.Context, token string) (*auth.Token, error) {
	client, err := firebaseClient()
	if err != nil {
		return nil, err
	}
	return client.authClient.VerifyIDToken(c, token)
}

// GetUser return the firebase user for the username
func GetUser(c context.Context, token string) (*auth.UserRecord, error) {
	client, err := firebaseClient()
	if err != nil {
		return nil, err
	}
	return client.authClient.GetUser(c, token)
}

// SubscribeToTopics create subscription topics for a user
func SubscribeToTopics(c context.Context, user *models.User, token string) error {
	client, err := firebaseClient()
	if err != nil {
		return err
	}
	var t string
	if user.IsSuperAdmin() {
		t = fmt.Sprintf("%s/superadmin", user.TenantID)
	}
	if user.IsAdmin() {
		t = fmt.Sprintf("%s/admin", user.TenantID)
	}
	if user.IsBackOffice() {
		t = fmt.Sprintf("%s/backoffice", user.TenantID)
	}
	if user.IsCustomer() {
		t = fmt.Sprintf("%s/customer/%s", user.TenantID, user.CustomerID.UUID.String())
	}
	if user.IsDriver() {
		t = fmt.Sprintf("%s/driver", user.TenantID)
	}
	var topics = []string{t}
	for _, t := range topics {

		r, err := client.messagingClient.SubscribeToTopic(c, []string{token}, t)
		if err != nil {
			log.Println(user.ID, " role=", user.Role, " subscription failed to topic", t, r)
			return err
		}
		log.Println(user.ID, " role=", user.Role, " subscribed to topic", t, r)
	}
	return nil
}

// UnSubscribeToTopics removes subscription topics for a user
func UnSubscribeToTopics(c context.Context, user *models.User, token string) error {
	client, err := firebaseClient()
	if err != nil {
		return err
	}
	topics := []string{fmt.Sprintf("%s/superadmin", user.TenantID), fmt.Sprintf("%s/admin", user.TenantID),
		fmt.Sprintf("%s/backoffice", user.TenantID), fmt.Sprintf("%s/driver", user.TenantID),
	}
	if user.IsCustomer() {
		topics = append(topics, fmt.Sprintf("%s/customer/%s", user.TenantID, user.CustomerID.UUID.String()))
	}
	for _, t := range topics {
		r, err := client.messagingClient.UnsubscribeFromTopic(c, []string{token}, t)
		if err != nil {
			log.Println(user.ID, " role=", user.Role, " unsubscription failed to topic", t, r)
			return err
		}
		log.Println(user.ID, " role=", user.Role, " unsubscribed to topic", t, r)
	}
	return nil
}
