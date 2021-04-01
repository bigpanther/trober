package grifts

import (
	"log"

	"github.com/bigpanther/trober/actions"
	"github.com/bigpanther/trober/firebase"

	"github.com/gobuffalo/buffalo"
)

func init() {
	f, err := firebase.NewFake()
	if err != nil {
		log.Fatalln("failed to connect to fake firebase", err)
	}
	buffalo.Grifts(actions.App(f))
}
