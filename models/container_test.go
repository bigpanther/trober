package models

import (
	"fmt"
	"testing"

	"github.com/gobuffalo/nulls"
)

func (ms *ModelSuite) Test_Container() {
	var tests = []struct {
		container                *Container
		expectedValidationErrors int
	}{
		{&Container{}, 3},
		{&Container{SerialNumber: "CANV0001"}, 2},
		{&Container{SerialNumber: "CANV0001", Type: ContainerTypeIncoming.String()}, 1},
		{&Container{SerialNumber: "CANV0001", Type: ContainerTypeIncoming.String(), Status: ContainerStatusAbandoned.String()}, 0},
		{&Container{Size: nulls.NewString("Invalid size")}, 4},
	}
	for i, test := range tests {
		ms.T().Run(fmt.Sprint(i), func(t *testing.T) {
			v, err := test.container.Validate(ms.DB)
			ms.Nil(err)
			ms.Equal(test.expectedValidationErrors, len(v.Errors))
		})
	}
}
