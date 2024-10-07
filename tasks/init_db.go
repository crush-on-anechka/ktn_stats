package tasks

import (
	"fmt"

	"github.com/crush-on-anechka/ktn_stats/db"
)

// InitDB creates SQlite database based on Data struct from models.go
func InitDB() error {
	db, err := db.NewSqliteDB()
	if err != nil {
		return fmt.Errorf("failed to establish connection with database: %w", err)
	}
	defer db.DB.Close()

	err = db.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	return nil
}
