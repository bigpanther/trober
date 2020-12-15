package main

import (
	"log"

	"github.com/bigpanther/trober/actions"
)

// main is the starting point for your Buffalo application.
// You can feel free and add to this `main` method, change
// what it does, etc...
// All we ask is that, at some point, you make sure to
// call `app.Serve()`, unless you don't want to start your
// application that is. :)
func main() {
	app := actions.App()
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}
