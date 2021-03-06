package firebase

//go:generate mockgen -destination ../actions/mock_firebase_test.go -package actions github.com/bigpanther/trober/firebase Firebase

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/messaging"
	"github.com/bigpanther/trober/models"
	"google.golang.org/api/option"
)

type Firebase interface {
	SetClaims(c context.Context, u *models.User) error
	GetUser(c context.Context, token string) (*auth.UserRecord, error)
	SendAll(c context.Context, messages []*messaging.Message) error
	SubscribeToTopics(c context.Context, user *models.User, token string) error
	UnSubscribeToTopics(c context.Context, user *models.User, token string) error
}
type firebaseSdkClient struct {
	authClient      *auth.Client
	messagingClient *messaging.Client
}

var errMissingToken = errors.New("missing token")

// New returns an instance of Firebase
func New() (Firebase, error) {
	var credsJSONEncoded = os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON_ENCODED")
	credJSON, err := base64.StdEncoding.DecodeString(credsJSONEncoded)
	if err != nil {
		return nil, err
	}
	opt := option.WithCredentialsJSON(credJSON)
	ctx := context.Background()
	client := &firebaseSdkClient{}
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
func (client *firebaseSdkClient) SetClaims(c context.Context, u *models.User) error {
	claims := map[string]interface{}{
		"bpTenantId":   u.TenantID.String(),
		"bpRole":       u.Role,
		"bpCustomerId": u.CustomerID.UUID.String(),
	}
	if !u.CustomerID.Valid {
		claims["bpCustomerId"] = "none"
	}
	err := client.authClient.SetCustomUserClaims(c, u.Username, claims)
	if err != nil {
		return err
	}
	return nil
}

// verifyIDToken return the auth token after verification
func (client *firebaseSdkClient) verifyIDToken(c context.Context, token string) (*auth.Token, error) {
	return client.authClient.VerifyIDToken(c, token)
}

// GetUser return the firebase user for the username
func (client *firebaseSdkClient) GetUser(c context.Context, token string) (*auth.UserRecord, error) {
	t, err := client.verifyIDToken(c, token)
	if err != nil {
		return nil, err
	}
	return client.authClient.GetUser(c, t.Subject)
}

// SendAll sends all messages to FCM topics
func (client *firebaseSdkClient) SendAll(c context.Context, messages []*messaging.Message) error {
	r, err := client.messagingClient.SendAll(c, messages)
	if err != nil {
		log.Println("error sending messages", err, r)
		return err
	}
	return nil
}

// SubscribeToTopics create subscription topics for a user
func (client *firebaseSdkClient) SubscribeToTopics(c context.Context, user *models.User, token string) error {
	if token == "" {
		return errMissingToken
	}

	var t = GetTopic(user)
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
func (client *firebaseSdkClient) UnSubscribeToTopics(c context.Context, user *models.User, token string) error {
	if token == "" {
		return errMissingToken
	}
	topics := []string{GetSuperAdminTopic(), GetAdminTopic(user),
		GetBackOfficeTopic(user), fmt.Sprintf("%s_driver_%s", user.TenantID, user.ID),
	}
	if user.IsCustomer() {
		topics = append(topics, fmt.Sprintf("%s_customer_%s", user.TenantID, user.CustomerID.UUID.String()))
	}
	for _, t := range topics {
		r, err := client.messagingClient.UnsubscribeFromTopic(c, []string{token}, t)
		if err != nil {
			log.Println(user.ID, " role=", user.Role, " unsubscription failed to topic", t, r)
			// Don't fail here
			continue
		}
		log.Println(user.ID, " role=", user.Role, " unsubscribed to topic", t, r)
	}
	return nil
}

// GetTopic returns the topic name for the user
func GetTopic(user *models.User) string {
	if user.IsSuperAdmin() {
		return GetSuperAdminTopic()
	}
	if user.IsAdmin() {
		return GetAdminTopic(user)
	}
	if user.IsBackOffice() {
		return GetBackOfficeTopic(user)
	}
	if user.IsCustomer() {
		return GetCustomerTopic(user.TenantID.String(), user.CustomerID.UUID.String())
	}
	if user.IsDriver() {
		return GetDriverTopic(user.TenantID.String(), user.ID.String())
	}
	return fmt.Sprintf("%s_none_%s", user.TenantID, user.ID)
}

// GetSuperAdminTopic returns the topic for superuser
func GetSuperAdminTopic() string {
	return fmt.Sprint("superadmin")
}

// GetCustomerTopic returns the topic for superuser
func GetCustomerTopic(tenantID string, customerID string) string {
	return fmt.Sprintf("%s_customer_%s", tenantID, customerID)
}

// GetDriverTopic returns the topic for superuser
func GetDriverTopic(tenantID string, driverID string) string {
	return fmt.Sprintf("%s_driver_%s", tenantID, driverID)
}

// GetAdminTopic returns the topic for the admin user
func GetAdminTopic(user *models.User) string {
	return fmt.Sprintf("%s_admin", user.TenantID)
}

// GetBackOfficeTopic returns the topic for the back office user
func GetBackOfficeTopic(user *models.User) string {
	return fmt.Sprintf("%s_backoffice", user.TenantID)
}
