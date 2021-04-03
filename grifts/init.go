package grifts

import (
	"log"

	"github.com/bigpanther/trober/actions"
	"github.com/bigpanther/trober/firebase"

	"github.com/gobuffalo/buffalo"
)

func init() {
	isProd := actions.ENV == "production"
	var (
		f   firebase.Firebase
		err error
	)
	if isProd {
		f, err = firebase.New()
	} else {
		f, err = firebase.NewFake()
	}
	if err != nil {
		log.Fatal("failed it initialize connection to firebase", err)
	}
	buffalo.Grifts(actions.App(f))
}
