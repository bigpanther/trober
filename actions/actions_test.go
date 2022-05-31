package actions

import (
	"os"
	"testing"

	"github.com/gobuffalo/suite/v4"
	"github.com/golang/mock/gomock"
)

type ActionSuite struct {
	*suite.Action
}

var mockFirebase *MockFirebase

func Test_ActionSuite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFirebase = NewMockFirebase(ctrl)
	action, err := suite.NewActionWithFixtures(App(mockFirebase), os.DirFS("../fixtures"))
	if err != nil {
		t.Fatal(err)
	}

	as := &ActionSuite{
		Action: action,
	}
	suite.Run(t, as)
}
