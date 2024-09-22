package tasks

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/crush-on-anechka/ktn_stats/config"
	"github.com/crush-on-anechka/ktn_stats/db"
	"github.com/crush-on-anechka/ktn_stats/sheetsclient"
	"github.com/crush-on-anechka/ktn_stats/utils"
)

// CheckFieldnames parses fieldnames from most recent spreadsheet and checks if they all are
// present in db.Data struct
func CheckFieldnames() error {
	client, err := sheetsclient.New(config.GreedyRequestTimeout)
	if err != nil {
		utils.HandleError(err, config.SheetsErrorMsg)
	}

	currentYear := time.Now().Year()
	yearAsStr := strconv.Itoa(currentYear)

	currentSpreadsheet := client.GetSpreadsheetByYear(yearAsStr)

	if currentSpreadsheet == nil {
		return config.ErrCurYearSpreadsheet
	}

	fieldnamesFromSheets := client.GetFieldnamesFromSpreadsheet(currentSpreadsheet)

	for _, field := range config.ExcludeFields {
		delete(fieldnamesFromSheets, field)
	}

	if !areFieldnamesPresentInModel(fieldnamesFromSheets) {
		return errors.New("spreadsheet contains new fields which are not present in database")
	}

	return nil
}

// areFieldnamesPresentInModel checks if all fieldnames from given map exist in db.Data struct
func areFieldnamesPresentInModel(fieldnamesFromSheets map[string]bool) bool {
	val := reflect.ValueOf(db.Data{})
	typ := val.Type()

	fieldnamesFromModel := make(map[string]bool)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tagValue := field.Tag.Get("fieldname")
		fieldnamesFromModel[tagValue] = true
	}

	for key := range fieldnamesFromSheets {
		if !fieldnamesFromModel[key] {
			fmt.Printf("missing key: %s\n", key)
			return false
		}
	}

	return true
}
