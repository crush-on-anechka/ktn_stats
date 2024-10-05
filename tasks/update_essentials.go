package tasks

import (
	"fmt"

	"github.com/crush-on-anechka/ktn_stats/db"
	"github.com/crush-on-anechka/ktn_stats/essentialshandler"
)

func UpdateEssentials() error {
	storage, err := db.NewSqliteDB()
	if err != nil {
		return fmt.Errorf("failed to establish connection with database: %w", err)
	}

	defer storage.DB.Close()

	dates, err := storage.GetDates()
	if err != nil {
		return fmt.Errorf("failed to fetch dates from db: %w", err)
	}

	essentialsHandler := essentialshandler.New(storage)

	for _, date := range dates {
		if err = essentialsHandler.UpdateEssentialsByDate(date); err != nil {
			return fmt.Errorf("failed to update essentials for date %s: %w", date, err)
		}
	}

	return nil
}
