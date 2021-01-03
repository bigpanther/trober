package grifts

import (
	"errors"
	"log"

	"github.com/bigpanther/trober/firebase"
	"github.com/bigpanther/trober/models"
	"github.com/markbates/grift/grift"
)

var _ = grift.Namespace("claims", func() {
	grift.Desc("reset", "Resets the custom user claims")
	grift.Add("reset", func(c *grift.Context) error {
		users := models.Users{}
		err := models.DB.All(&users)
		if err != nil {
			return err
		}
		for _, u := range users {
			if err := firebase.SetClaims(c, &u); err != nil {
				return err
			}
		}
		return nil
	})

	grift.Desc("show <id>", "Show the custom user claims")
	grift.Add("show", func(c *grift.Context) error {
		var args = c.Args
		if len(args) < 1 {
			return errors.New("missing user id")
		}
		user := models.User{}
		userID := args[0]
		log.Println("fetching claims for user id ", userID)
		err := models.DB.Find(&user, userID)
		if err != nil {
			return err
		}

		u, err := firebase.GetUser(c, user.Username)
		if err != nil {
			return err
		}
		log.Println(u.CustomClaims)
		return nil
	})
})
