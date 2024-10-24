package main

import (
	"flag"
	"log"
	"time"

	"github.com/crush-on-anechka/ktn_stats/config"
	"github.com/crush-on-anechka/ktn_stats/messagesender"
)

func initFlags() (*bool, *bool, map[string]*bool, *string) {
	taskMode := flag.Bool("task", false, "Run cron task")
	webMode := flag.Bool("web", false, "Run as web server")

	taskFlags := map[string]*bool{
		"init_db":           flag.Bool("init_db", false, "Initialize database"),
		"check_fieldnames":  flag.Bool("check_fieldnames", false, "Check fields completeness"),
		"store_by_year":     flag.Bool("store_by_year", false, "Fetch and store spreadsheet"),
		"store_latest":      flag.Bool("store_latest", false, "Fetch and store latest spreadsheet"),
		"store_all":         flag.Bool("store_all", false, "Fetch and store all spreadsheets"),
		"update_essentials": flag.Bool("update_essentials", false, "Re-process essential fields"),
	}

	year := flag.String("year", "", "Specify year for storing spreadsheet data")

	flag.Parse()

	return taskMode, webMode, taskFlags, year
}

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
