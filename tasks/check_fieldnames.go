package tasks

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/crush-on-anechka/ktn_stats/config"
	"github.com/crush-on-anechka/ktn_stats/db"
	"github.com/crush-on-anechka/ktn_stats/sheetsclient"
)

// CheckFieldnames parses fieldnames from most recent spreadsheet and checks if they all are
// present in db.Data struct
func CheckFieldnames() error {
	client, err := sheetsclient.New(config.GreedyRequestTimeout)
	if err != nil {
		return fmt.Errorf("failed to create Sheets client: %w", err)
	}

	currentYear := time.Now().Year()
	yearAsStr := strconv.Itoa(currentYear)

	currentSpreadsheet, err := client.GetSpreadsheetByYear(yearAsStr)
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet by year %s: %w", yearAsStr, err)
	}

	fieldnamesFromSheets, err := client.GetFieldnamesFromSpreadsheet(currentSpreadsheet)
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet fieldnames: %w", err)
	}

	for _, field := range config.ExcludeFields {
		delete(fieldnamesFromSheets, field)
	}

	if err := fieldnamesPresentInModelCheck(fieldnamesFromSheets); err != nil {
		return fmt.Errorf("spreadsheet %s contains fields which are not present in database: %w",
			currentSpreadsheet.Properties.Title, err)
	}
	return nil
}

// fieldnamesPresentInModelCheck checks if all fieldnames from given map exist in db.Data struct
func fieldnamesPresentInModelCheck(fieldnamesFromSheets map[string]bool) error {
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
			return fmt.Errorf("key %s is not found in Data model", key)
		}
	}
	return nil
}
