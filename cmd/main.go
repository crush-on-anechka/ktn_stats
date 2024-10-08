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
	taskMode, webMode, taskFlags := initFlags()

	botToken := config.Envs.TelegramToken
	chatID := int64(config.Envs.TelegramChatID)
	sender, err := messagesender.New(botToken, chatID)
	if err != nil {
		log.Fatal(err)
	}

	if *webMode {
		startServer(sender)
	} else if *taskMode {
		runTask(taskFlags, sender)
	} else {
		log.Println("No mode specified. Use --task or --web")
	}
}

func runTask(taskFlags map[string]*bool, sender *messagesender.Sender) {
	switch {
	case *taskFlags["init_db"]:
		err := tasks.InitDB()
		handleError(err, sender, "Failed to initialize database")
		handleSuccess(sender, "Database was successfully initialized")

	case *taskFlags["check_fieldnames"]:
		err := tasks.CheckFieldnames()
		handleError(err, sender, "Failed to check fieldnames")
		handleSuccess(sender, "Fieldnames check: OK")

	case *taskFlags["store_all"]:
		err := tasks.StoreAllSpreadsheets()
		handleError(err, sender, "Failed to store spreadsheets data")
		handleSuccess(sender, "Spreadsheets data was successfully stored")

	case *taskFlags["store_latest"]:
		err := tasks.StoreLatestSpreadsheet()
		handleError(err, sender, "Failed to store latest spreadsheet data")
		handleSuccess(sender, "Latest spreadsheet data was successfully stored")

	case *taskFlags["update_essentials"]:
		err := tasks.UpdateEssentials()
		handleError(err, sender, "Failed to update essentials")
		handleSuccess(sender, "Essential words and phrases were successfully updated")

	default:
		fmt.Println("No task specified. Available flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func startServer(sender *messagesender.Sender) {
	// http.HandleFunc("/fetch", fetchDataFromDB)
	// log.Println("Starting HTTP server on :8080")

	// err := http.ListenAndServe(":8080", nil)
	// handleError(err, sender, "Failed to start HTTP server")
}

// func fetchDataFromDB(w http.ResponseWriter, r *http.Request) {
// 	name := r.URL.Query().Get("name")

// 	// Write response
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	fmt.Fprintf(w, `{"message": "Hello, %s!"}`, name)
// }
