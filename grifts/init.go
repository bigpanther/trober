package grifts

import (
	"github.com/bigpanther/trober/actions"

	"github.com/gobuffalo/buffalo"
)

func init() {
	buffalo.Grifts(actions.App())
}
