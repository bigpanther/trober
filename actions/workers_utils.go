package actions

import "github.com/gobuffalo/buffalo/worker"

func sendNotificationsAsync(topics []string, messageTitle string, messageBody string, data map[string]string) {
	app.Worker.Perform(worker.Job{
		Queue:   "default",
		Handler: "sendNotifications",
		Args: worker.Args{
			"topics":        topics,
			"message.title": messageTitle,
			"message.body":  messageBody,
			"message.data":  data,
		},
	})
}
