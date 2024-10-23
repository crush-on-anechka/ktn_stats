package tasks

import (
	"fmt"
	"strconv"

	"github.com/crush-on-anechka/ktn_stats/config"
	"github.com/crush-on-anechka/ktn_stats/db"
	"github.com/crush-on-anechka/ktn_stats/essentialshandler"
	"github.com/crush-on-anechka/ktn_stats/sheetshandler"
)

func StoreSpreadsheet(year string) error {
	storage, err := db.NewSqliteDB()
	if err != nil {
		return fmt.Errorf("failed to establish connection with database: %w", err)
	}
	defer storage.DB.Close()

	essentialsHandler := essentialshandler.New(storage)

	sheetsHandler, err := sheetshandler.New(storage, config.GreedyRequestTimeout, essentialsHandler)
	if err != nil {
		return fmt.Errorf("failed to create sheetshandler: %w", err)
	}

	yearAsInt, err := strconv.Atoi(year)
	if err != nil {
		return fmt.Errorf("failed to parse year: %w", err)
	}

	return sheetsHandler.StoreSpreadsheetByYear(yearAsInt)
}
