package tasks

import (
	"github.com/crush-on-anechka/ktn_stats/db"
	"github.com/crush-on-anechka/ktn_stats/utils"
)

// InitDB creates SQlite database based on Data struct from models.go
func InitDB() {
	db, err := db.NewSqliteDB()
	if err != nil {
		utils.HandleError(err, "Failed to establish connection with database")
	}

	defer db.DB.Close()

	err = db.Init()
	if err != nil {
		utils.HandleError(err, "Failed to initialize database")
	}
}
