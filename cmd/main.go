package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/crush-on-anechka/ktn_stats/tasks"
	"github.com/crush-on-anechka/ktn_stats/utils"
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
		"Fetches and stores data from all spreadsheets",
	)
	StoreLatestFlag := flag.Bool(
		"store_latest",
		false,
		"Fetches and stores data from most recent spreadsheet",
	)

	flag.Parse()

	if *InitDBFlag {
		tasks.InitDB()
	} else if *CheckFieldnamesFlag {
		err := tasks.CheckFieldnames()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Fieldnames check: OK")
		}
	} else if *StoreAllFlag {
		err := tasks.StoreAllSpreadsheets()
		if err != nil {
			utils.HandleError(err, "failed to store spreadsheets data")
		} else {
			fmt.Println("Spreadsheets data were successfully stored")
		}
	} else if *StoreLatestFlag {
		err := tasks.StoreLatestSpreadsheet()
		if err != nil {
			utils.HandleError(err, "failed to store latest spreadsheet data")
		} else {
			fmt.Println("Latest spreadsheet data was successfully stored")
		}
	} else {
		fmt.Println("No task specified. Available flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}
}
