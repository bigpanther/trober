package actions

import (
	"context"
	"log"

	"firebase.google.com/go/v4/messaging"
	"github.com/bigpanther/trober/firebase"
	"github.com/gobuffalo/buffalo/worker"
)

func sendNotifications(args worker.Args) error {
	var tos = args["to"].([]string)
	msgTitle := args["message.title"].(string)
	msgBody := args["message.body"].(string)
	msgData := args["message.data"].(map[string]string)
	var messages []*messaging.Message
	for _, to := range tos {
		message := &messaging.Message{
			Data: msgData,
			Notification: &messaging.Notification{
				Title: msgTitle,
				Body:  msgBody,
			},
			Topic: to,
		}
		messages = append(messages, message)
	}
	// TODO: Add a timeout here
	return firebase.SendAll(context.Background(), messages)
}
func testWorker(args worker.Args) error {
	log.Println(args)
	return nil
}
