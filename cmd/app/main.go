package main

import (
	"log"

	"github.com/bigpanther/trober/actions"
	"github.com/bigpanther/trober/firebase"
)

// main is the starting point for your Buffalo application.
// You can feel free and add to this `main` method, change
// what it does, etc...
// All we ask is that, at some point, you make sure to
// call `app.Serve()`, unless you don't want to start your
// application that is. :)
func main() {
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
	app := actions.App(f)
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}
