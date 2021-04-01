package actions

import (
	"testing"

	"github.com/bigpanther/trober/firebase"
	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/suite/v3"
)

type ActionSuite struct {
	*suite.Action
}

func Test_ActionSuite(t *testing.T) {
	// TODO: Validate sendNotification callback by mocking the Firebase interface
	f, err := firebase.NewFake()
	if err != nil {
		t.Fatalf("error connecting to firebase: %v\n", err)
	}
	action, err := suite.NewActionWithFixtures(App(f), packr.New("Test_ActionSuite", "../fixtures"))
	if err != nil {
		t.Fatal(err)
	}

	as := &ActionSuite{
		Action: action,
	}
	suite.Run(t, as)
}
