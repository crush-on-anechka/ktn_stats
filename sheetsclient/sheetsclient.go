package sheetsclient

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/crush-on-anechka/ktn_stats/config"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SheetsClient struct {
	Service        *sheets.Service
	spreadsheetIDs []string
	RequestTimeout time.Duration
}

func New(requestTimeout time.Duration) (*SheetsClient, error) {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(config.CredentialsFile))
	if err != nil {
		return nil, err
	}
	sheetsClient := &SheetsClient{
		Service:        srv,
		spreadsheetIDs: config.Envs.SpreadsheetIDs,
		RequestTimeout: requestTimeout,
	}
	return sheetsClient, nil
}

func (client *SheetsClient) GetSpreadsheetByID(spreadsheetID string) (*sheets.Spreadsheet, error) {
	spreadsheet, err := client.Service.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return nil, err
	}
	return spreadsheet, nil
}

func (client *SheetsClient) GetSpreadsheetByYear(year string) *sheets.Spreadsheet {
	for _, spreadsheetID := range client.spreadsheetIDs {
		spreadsheet, err := client.Service.Spreadsheets.Get(spreadsheetID).Do()
		if err != nil {
			return nil
		}

		spreadsheetYear := ExtractYearFromTitle(spreadsheet.Properties.Title)

		if spreadsheetYear == year {
			return spreadsheet
		}
	}
	return nil
}

func (client *SheetsClient) GetSpreadsheets() ([]*sheets.Spreadsheet, error) {
	spreadsheets := []*sheets.Spreadsheet{}

	for _, spreadsheetID := range client.spreadsheetIDs {
		spreadsheet, err := client.Service.Spreadsheets.Get(spreadsheetID).Do()
		if err != nil {
			return nil, err
		}
		spreadsheets = append(spreadsheets, spreadsheet)
	}

	return spreadsheets, nil
}

// getFieldnamesFromSpreadsheet parses all existing column (field) names from every sheet
// in a specified spreadsheet
func (client *SheetsClient) GetFieldnamesFromSpreadsheet(
	spreadsheet *sheets.Spreadsheet) map[string]bool {

	fieldnames := make(map[string]bool)

	for _, sheet := range spreadsheet.Sheets {
		time.Sleep(client.RequestTimeout)
		sheetName := sheet.Properties.Title

		if !config.DatePatternRegex.MatchString(sheetName) {
			continue
		}

		readRange := fmt.Sprintf("%s!%s", sheetName, config.Envs.SheetParseRange)

		resp, err := client.Service.Spreadsheets.Values.Get(
			spreadsheet.SpreadsheetId, readRange).Do()
		if err != nil {
			log.Fatalf("Unable to retrieve data from sheet: %v", err)
		}

		for _, row := range resp.Values {
			for _, cell := range row {
				cellStr, ok := cell.(string)
				if !ok {
					fmt.Println("Type assertion failed. The interface does not contain a string.")
					continue
				}

				if !IsNumeric(cellStr) {
					fieldnames[cellStr] = true
				}
			}
			break
		}
	}

	return fieldnames
}

func IsNumeric(value string) bool {
	cleanedValue := strings.ReplaceAll(value, "\u00A0", "")
	_, err := strconv.Atoi(cleanedValue)
	return err == nil
}

func ExtractYearFromTitle(input string) string {
	re := regexp.MustCompile(`\b\d{4}\b`)

	year := re.FindString(input)

	return year
}
