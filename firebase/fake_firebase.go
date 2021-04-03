package firebase

import (
	"context"

	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/messaging"
	"github.com/bigpanther/trober/models"
)

type fakefirebase struct {
}

// NewFake returns a fake/noop instance of Firebase
func NewFake() (Firebase, error) {
	return &fakefirebase{}, nil
}

// SetClaims sets the custom claims for the user
func (client *fakefirebase) SetClaims(c context.Context, u *models.User) error {

	return nil
}

// GetUser return the firebase user for the username
func (client *fakefirebase) GetUser(c context.Context, token string) (*auth.UserRecord, error) {
	return &auth.UserRecord{}, nil
}

// SendAll sends all messages to FCM topics
func (client *fakefirebase) SendAll(c context.Context, messages []*messaging.Message) error {

	return nil
}

// SubscribeToTopics create subscription topics for a user
func (client *fakefirebase) SubscribeToTopics(c context.Context, user *models.User, token string) error {

	return nil
}

// UnSubscribeToTopics removes subscription topics for a user
func (client *fakefirebase) UnSubscribeToTopics(c context.Context, user *models.User, token string) error {

	return nil
}
