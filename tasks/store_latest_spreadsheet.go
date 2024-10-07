package tasks

import (
	"fmt"
	"time"

	"github.com/crush-on-anechka/ktn_stats/config"
	"github.com/crush-on-anechka/ktn_stats/db"
	"github.com/crush-on-anechka/ktn_stats/essentialshandler"
	"github.com/crush-on-anechka/ktn_stats/sheetshandler"
)

func StoreLatestSpreadsheet() error {
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

	currentYear := time.Now().Year()

	return sheetsHandler.StoreSpreadsheetByYear(currentYear)
}
