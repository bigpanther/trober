package models

import (
	"fmt"
	"testing"
)

func (ms *ModelSuite) Test_Terminal() {
	var tests = []struct {
		terminal                 *Terminal
		expectedValidationErrors int
	}{
		{&Terminal{}, 2},
		{&Terminal{Type: "invalid"}, 2},
		{&Terminal{Name: "some name"}, 1},
		{&Terminal{Name: "some name", Type: TerminalTypeAirport.String()}, 0},
	}
	for i, test := range tests {
		ms.T().Run(fmt.Sprint(i), func(t *testing.T) {
			v, err := test.terminal.Validate(ms.DB)
			ms.Nil(err)
			ms.Equal(test.expectedValidationErrors, len(v.Errors))
		})
	}
}
