package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/crush-on-anechka/ktn_stats/config"
	"github.com/crush-on-anechka/ktn_stats/db"
	"github.com/crush-on-anechka/ktn_stats/messagesender"
	"github.com/crush-on-anechka/ktn_stats/tasks"
	"github.com/gorilla/mux"
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
	db, err := db.NewSqliteDB()
	handleError(err, sender, "Failed to establish connection with database")
	defer db.DB.Close()

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	r.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		fetchDataFromDB(w, r, db)
	})

	handleSuccess(sender, fmt.Sprintf("Starting HTTP server on :%v", config.Envs.APIPort))

	err = http.ListenAndServe(fmt.Sprintf(":%v", config.Envs.APIPort), r)
	handleError(err, sender, "Failed to start HTTP server")
}

func fetchDataFromDB(w http.ResponseWriter, r *http.Request, storage *db.SqliteDB) {
	query := r.URL.Query().Get("search")
	wholePhrase := r.URL.Query().Get("wholePhrase") != ""

	w.Header().Set("Content-Type", "application/json")

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// TODO: set these up
	page := 1
	limit := 10

	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}
	if limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}

	searchType := r.URL.Query().Get("searchType")

	var result []db.Data
	var err error

	if searchType == "byInscription" {
		result, err = storage.GetOrdersBySearch(query, wholePhrase)
	} else if searchType == "byCustomer" {
		result, err = storage.GetOrdersByCustomer(query)
	}
	if err != nil {
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
	}

	start := (page - 1) * limit
	end := start + limit

	if start > len(result) {
		http.Error(w, "Page out of range", http.StatusBadRequest)
		return
	}
	if end > len(result) {
		end = len(result)
	}

	paginatedResult := result[start:end]

	if err := json.NewEncoder(w).Encode(paginatedResult); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}
