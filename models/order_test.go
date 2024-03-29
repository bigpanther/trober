package models

import (
	"fmt"
	"testing"
)

func (ms *ModelSuite) TestOrder() {
	var tests = []struct {
		order                    *Order
		expectedValidationErrors int
	}{
		{&Order{}, 3},
		{&Order{Status: "invalid"}, 3},
		{&Order{SerialNumber: "ORD0001"}, 2},
		{&Order{SerialNumber: "ORD0001", Status: OrderStatusAccepted.String()}, 1},
		{&Order{SerialNumber: "ORD0001", Status: OrderStatusAccepted.String(), Type: ShipmentTypeInbound.String()}, 0},
	}
	for i, test := range tests {
		ms.T().Run(fmt.Sprint(i), func(t *testing.T) {
			v, err := test.order.Validate(ms.DB)
			ms.Nil(err)
			ms.Equal(test.expectedValidationErrors, len(v.Errors))
		})
	}
}
