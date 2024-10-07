package tasks

import (
	"fmt"
	"time"

	"github.com/crush-on-anechka/ktn_stats/config"
	"github.com/crush-on-anechka/ktn_stats/db"
	"github.com/crush-on-anechka/ktn_stats/essentialshandler"
	"github.com/crush-on-anechka/ktn_stats/sheetshandler"
)

func StoreAllSpreadsheets() error {
	storage, err := db.NewSqliteDB()
	if err != nil {
		return fmt.Errorf("failed to establish connection with database: %w", err)
	}
	defer storage.DB.Close()

	essentialsHandler := essentialshandler.New(storage)

	sheetsHandler, err := sheetshandler.New(storage, config.SafeRequestTimeout, essentialsHandler)
	if err != nil {
		return fmt.Errorf("failed to create sheetshandler: %w", err)
	}

	currentYear := time.Now().Year()

	for year := config.StartYear; year <= currentYear; year++ {
		err := sheetsHandler.StoreSpreadsheetByYear(year)
		if err != nil {
			return fmt.Errorf("failed to store %v spreadsheet: %w", currentYear, err)
		}
	}

	return nil
}
