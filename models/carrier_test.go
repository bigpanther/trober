package models

import (
	"fmt"
	"testing"
)

func (ms *ModelSuite) Test_Carrier() {
	var tests = []struct {
		carrier                  *Carrier
		expectedValidationErrors int
	}{
		{&Carrier{}, 2},
		{&Carrier{Type: "invalid"}, 2},
		{&Carrier{Name: "some name"}, 1},
		{&Carrier{Name: "some name", Type: CarrierTypeAir.String()}, 0},
	}
	for i, test := range tests {
		ms.T().Run(fmt.Sprint(i), func(t *testing.T) {
			v, err := test.carrier.Validate(ms.DB)
			ms.Nil(err)
			ms.Equal(test.expectedValidationErrors, len(v.Errors))
		})
	}
}
