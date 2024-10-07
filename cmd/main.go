package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/crush-on-anechka/ktn_stats/config"
	"github.com/crush-on-anechka/ktn_stats/messagesender"
	"github.com/crush-on-anechka/ktn_stats/tasks"
)

func main() {
	InitDBFlag := flag.Bool("init_db", false, "Initialize database")
	CheckFieldnamesFlag := flag.Bool(
		"check_fields",
		false,
		"Check if all fieldnames from most recent spreadsheet are present in db",
	)
	StoreAllFlag := flag.Bool(
		"store_all",
		false,
		"Fetch and store data from all spreadsheets",
	)
	StoreLatestFlag := flag.Bool(
		"store_latest",
		false,
		"Fetch and store data from most recent spreadsheet",
	)
	UpdateEssentialsFlag := flag.Bool(
		"update_essentials",
		false,
		"Force re-process inscriptions and update essentials",
	)

	flag.Parse()

	botToken := config.Envs.TelegramToken
	chatID := int64(config.Envs.TelegramChatID)
	sender, err := messagesender.New(botToken, chatID)
	if err != nil {
		log.Fatal(err)
	}

	switch {
	case *InitDBFlag:
		err := tasks.InitDB()
		handleError(err, sender, "Failed to initialize database")
		handleSuccess(sender, "Database was successfully initialized")

	case *CheckFieldnamesFlag:
		err := tasks.CheckFieldnames()
		handleError(err, sender, "Failed to check fieldnames")
		handleSuccess(sender, "Fieldnames check: OK")

	case *StoreAllFlag:
		err := tasks.StoreAllSpreadsheets()
		handleError(err, sender, "Failed to store spreadsheets data")
		handleSuccess(sender, "Spreadsheets data was successfully stored")

	case *StoreLatestFlag:
		err := tasks.StoreLatestSpreadsheet()
		handleError(err, sender, "Failed to store latest spreadsheet data")
		handleSuccess(sender, "Latest spreadsheet data was successfully stored")

	case *UpdateEssentialsFlag:
		err := tasks.UpdateEssentials()
		handleError(err, sender, "Failed to update essentials")
		handleSuccess(sender, "Essential words and phrases were successfully updated")

	default:
		fmt.Println("No task specified. Available flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}
}
