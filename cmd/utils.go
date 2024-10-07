package main

import (
	"log"
	"time"

	"github.com/crush-on-anechka/ktn_stats/config"
	"github.com/crush-on-anechka/ktn_stats/messagesender"
)

func handleError(err error, sender *messagesender.Sender, message string) {
	if err != nil {
		errSender := sender.SendMessageToTelegramBot(message)
		if errSender != nil {
			log.Println("Failed to send message to Telegram:", errSender)
		}
		log.Fatal(err)
	}
}

func handleSuccess(sender *messagesender.Sender, message string) {
	log.Println(message)
	weeklyCheck(sender, message)
}

func weeklyCheck(sender *messagesender.Sender, message string) {
	today := time.Now().Weekday()
	currentHour := time.Now().Hour()

	if today == config.WeeklyCheckWeekday &&
		currentHour >= config.WeeklyCheckHourFrom &&
		currentHour < config.WeeklyCheckHourTo {

		message = "Weekly check!\n" + message
		errSender := sender.SendMessageToTelegramBot(message)

		if errSender != nil {
			log.Println("Failed to send message to Telegram:", errSender)
		}
	}
}
