package tasks

import (
	"time"

	"github.com/crush-on-anechka/ktn_stats/config"
	"github.com/crush-on-anechka/ktn_stats/sheetshandler"
)

// весь капс
// все символы, если есть не кириллица

func StoreLatestSpreadsheet() error {
	handler := sheetshandler.New(config.GreedyRequestTimeout)

	defer handler.CloseDBconnection()

	currentYear := time.Now().Year()

	return handler.StoreSpreadsheetByYear(currentYear)
}
