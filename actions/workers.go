package actions

import (
	"log"

	"firebase.google.com/go/v4/messaging"
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
			Token: to,
		}
		messages = append(messages, message)
	}

	_, err := client.messagingClient.SendAll(app.Context, messages)
	if err != nil {
		log.Println("error sending message", err)
	}
	return nil
}
func testWorker(args worker.Args) error {
	log.Println(args)
	return nil
}
