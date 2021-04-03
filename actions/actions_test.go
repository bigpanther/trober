package actions

import (
	"testing"

	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/suite/v3"
	"github.com/golang/mock/gomock"
)

type ActionSuite struct {
	*suite.Action
}

var mockFirebase *MockFirebase

func Test_ActionSuite(t *testing.T) {
	// TODO: Validate sendNotification callback by mocking the Firebase interface
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFirebase = NewMockFirebase(ctrl)
	action, err := suite.NewActionWithFixtures(App(mockFirebase), packr.New("Test_ActionSuite", "../fixtures"))
	if err != nil {
		t.Fatal(err)
	}

	as := &ActionSuite{
		Action: action,
	}
	suite.Run(t, as)
}
