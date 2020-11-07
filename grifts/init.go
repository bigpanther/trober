package grifts

import (
	"github.com/shipanther/trober/actions"

	"github.com/gobuffalo/buffalo"
)

func init() {
	buffalo.Grifts(actions.App())
}
