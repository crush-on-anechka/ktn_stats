package tasks

import (
	"time"

	"github.com/crush-on-anechka/ktn_stats/config"
	"github.com/crush-on-anechka/ktn_stats/sheetshandler"
)

func StoreAllSpreadsheets() error {
	handler := sheetshandler.New(config.SafeRequestTimeout)

	defer handler.CloseDBconnection()

	currentYear := time.Now().Year()

	for year := config.StartYear; year <= currentYear; year++ {
		err := handler.StoreSpreadsheetByYear(year)
		if err != nil {
			return err
		}
	}
	return nil
}
