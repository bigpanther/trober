package models

import (
	"fmt"
	"testing"

	"github.com/gobuffalo/nulls"
)

func (ms *ModelSuite) Test_Shipment() {
	var tests = []struct {
		shipment                 *Shipment
		expectedValidationErrors int
	}{
		{&Shipment{}, 3},
		{&Shipment{SerialNumber: "CANV0001"}, 2},
		{&Shipment{SerialNumber: "CANV0001", Type: ShipmentTypeInbound.String()}, 1},
		{&Shipment{SerialNumber: "CANV0001", Type: ShipmentTypeInbound.String(), Status: ShipmentStatusDelivered.String()}, 0},
		{&Shipment{Size: nulls.NewString("Invalid size")}, 4},
	}
	for i, test := range tests {
		ms.T().Run(fmt.Sprint(i), func(t *testing.T) {
			v, err := test.shipment.Validate(ms.DB)
			ms.Nil(err)
			ms.Equal(test.expectedValidationErrors, len(v.Errors))
		})
	}
}
