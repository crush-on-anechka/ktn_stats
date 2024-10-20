package sheetshandler

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/crush-on-anechka/ktn_stats/config"
	"github.com/crush-on-anechka/ktn_stats/db"
	"github.com/crush-on-anechka/ktn_stats/essentialshandler"
	"github.com/crush-on-anechka/ktn_stats/sheetsclient"
	"google.golang.org/api/sheets/v4"
)

type SheetsHandler struct {
	client            *sheetsclient.SheetsClient
	storage           *db.SqliteDB
	essentialsHandler *essentialshandler.EssentialsHandler
}

func New(storage *db.SqliteDB,
	requestTimeout time.Duration,
	essentialsHandler *essentialshandler.EssentialsHandler,
) (*SheetsHandler, error) {

	client, err := sheetsclient.New(requestTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create Sheets client: %w", err)
	}

	return &SheetsHandler{
		client:            client,
		storage:           storage,
		essentialsHandler: essentialsHandler,
	}, nil
}

func (handler *SheetsHandler) StoreSpreadsheetByYear(inputYear int) error {
	inputYearAsStr := strconv.Itoa(inputYear)
	spreadsheet, err := handler.client.GetSpreadsheetByYear(inputYearAsStr)
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet by year %s: %w", inputYearAsStr, err)
	}

	for _, sheet := range spreadsheet.Sheets {
		time.Sleep(handler.client.RequestTimeout)

		sheetName := sheet.Properties.Title
		dateFromSheetName := config.DatePatternRegex.FindString(sheetName)

		if sheetName == config.SheetNameAvailability {
			dateFromSheetName = config.SheetAvailabilityDate
		}
		if sheetName == config.SheetNameUrgentOrders {
			dateFromSheetName = config.SheetUrgentOrdersDate
		}
		if dateFromSheetName == "" {
			continue
		}

		readRange := fmt.Sprintf("%s!%s", sheetName, config.Envs.SheetParseRange)
		resp, err := handler.client.Service.Spreadsheets.Values.Get(
			spreadsheet.SpreadsheetId, readRange).Do()
		if err != nil {
			return fmt.Errorf(
				"failed to retrieve data from, sheet %s (%v): %w", sheetName, inputYear, err)
		}

		sheetHash, err := GenerateHash(resp.Values)
		if err != nil {
			return fmt.Errorf(
				"failed to generate hash for sheet %s (%v): %w", sheetName, inputYear, err)
		}

		date := SerializeDate(dateFromSheetName, inputYearAsStr)

		tx, err := handler.storage.BeginTransaction()
		if err != nil {
			return err
		}

		defer func() {
			if p := recover(); p != nil {
				tx.Rollback()
				panic(p)
			} else if err != nil {
				tx.Rollback()
			}
		}()

		storedHash, err := handler.storage.GetHash(date)
		if err != nil {
			if errors.Is(err, config.ErrNoRecordFound) {
				handler.storage.CreateHashWithTx(tx, date, sheetHash)
			} else {
				return fmt.Errorf("failed to retrieve hash for date %s: %w", date, err)
			}
		}

		if sheetHash == storedHash {
			continue
		}

		if err = handler.processSheet(
			tx, sheet, sheetHash, date, spreadsheet.SpreadsheetId, resp.Values); err != nil {
			return fmt.Errorf("failed to process sheet: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		if err := handler.essentialsHandler.UpdateEssentialsByDate(date); err != nil {
			return err
		}

		log.Printf("Successsfuly stored data for %v\n", date)
	}

	return nil
}

func (handler *SheetsHandler) processSheet(
	tx *sql.Tx, sheet *sheets.Sheet, sheetHash, date, spreadsheetId string, values [][]interface{},
) error {

	handler.storage.DeleteDataByDateWithTx(tx, date)

	handler.storage.UpdateHashWithTx(tx, date, sheetHash)

	fieldnamesSlice := []string{}
	linkColumnExists := false
	dataToBeStored := []*db.Data{}

	mergedCells := getMergedCells(sheet)

	for rowIdx, row := range values {
		curRowData := processRow(rowIdx, row, &fieldnamesSlice, &linkColumnExists)

		_, merged := mergedCells[rowIdx]
		if merged && len(dataToBeStored) > 0 {
			curRowData["Ссылка"] = dataToBeStored[len(dataToBeStored)-1].CustomerLink
		} else if curRowData["Ссылка"] == "" {
			if curRowData["Сумма"] != "" {
				log.Printf("Link is missing in an entry with not-null sum: %v, line %v\n",
					date, rowIdx+1)
			} else {
				continue
			}
		}

		sheetID := sheet.Properties.SheetId
		rowNumber := rowIdx + 1
		orderLink := fmt.Sprintf(
			"https://docs.google.com/spreadsheets/d/%s/edit?gid=%v#gid=%v&range=%v:%v",
			spreadsheetId, sheetID, sheetID, rowNumber, rowNumber)

		NewDataInstance := &db.Data{
			Date:      date,
			RowNumber: rowNumber,
			IsMerged:  merged,
			OrderLink: orderLink,
		}

		if err := PopulateDataStructFromMap(NewDataInstance, curRowData); err != nil {
			return fmt.Errorf(
				"failed to convert map to Data struct: date %s, rowIdx: %v: %w",
				date, rowIdx, err,
			)
		}

		handleSearchField(NewDataInstance)

		dataToBeStored = append(dataToBeStored, NewDataInstance)
	}

	if err := handler.storage.BulkInsertDataWithTx(tx, dataToBeStored); err != nil {
		return fmt.Errorf("failed to perform bulk insert: %w", err)
	}

	return nil
}

func processRow(
	rowIdx int, row []interface{}, fieldnamesSlice *[]string, linkColumnExists *bool,
) map[string]string {

	curRowData := make(map[string]string)
	curRowData["Ссылка"] = ""

	for colIdx, cell := range row {
		if rowIdx > 0 && colIdx >= len(*fieldnamesSlice) {
			break
		}

		cellStr, ok := cell.(string)
		if !ok {
			log.Println(`Type assertion failed. The interface does
							not contain a string.`)
			continue
		}

		if rowIdx == 0 {
			if cellStr == "Ссылка" {
				*linkColumnExists = true
			}
			*fieldnamesSlice = append(*fieldnamesSlice, cellStr)
			continue
		}

		if !*linkColumnExists && colIdx == 0 {
			curRowData["Ссылка"] = cellStr
		} else if cellStr != "" {
			curRowData[(*fieldnamesSlice)[colIdx]] = cellStr
		}
	}

	return curRowData
}

func getMergedCells(sheet *sheets.Sheet) map[int]bool {
	mergedCells := make(map[int]bool)

	for _, mergeRange := range sheet.Merges {
		if mergeRange.StartColumnIndex > 1 {
			continue
		}
		for row := mergeRange.StartRowIndex + 1; row < mergeRange.EndRowIndex; row++ {
			mergedCells[int(row)] = true
		}
	}

	return mergedCells
}

func GenerateHash(data [][]interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal sheet data: %w", err)
	}

	hash := sha256.New()
	hash.Write([]byte(jsonData))
	hashAsString := hex.EncodeToString(hash.Sum(nil))

	return hashAsString, nil
}

func SerializeDate(sheetName, year string) string {
	parts := strings.Split(sheetName, ".")
	day := fmt.Sprintf("%02s", parts[0])
	month := fmt.Sprintf("%02s", parts[1])
	date := year + "." + month + "." + day
	return date
}

func PopulateDataStructFromMap(data *db.Data, values map[string]string) error {
	v := reflect.ValueOf(data).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		tag := field.Tag.Get("fieldname")
		value, exists := values[tag]
		if !exists || !fieldValue.CanSet() {
			continue
		}

		switch fieldValue.Kind() {
		case reflect.Int:
			intValue, err := strconv.Atoi(value)
			if err == nil {
				fieldValue.SetInt(int64(intValue))
			} else {
				floatValue, err := strconv.ParseFloat(strings.Replace(value, ",", ".", 1), 64)
				if err == nil {
					fieldValue.SetInt(int64(math.Round(floatValue)))
				} else {
					fieldValue.SetInt(0)
				}
			}

		case reflect.String:
			if field.Name == "Type" {
				value = strings.ToUpper(value)
			}
			fieldValue.SetString(value)

		default:
			return fmt.Errorf("unsupported field type %s for tag %s", fieldValue.Kind(), tag)
		}
	}
	return nil
}

func handleSearchField(NewDataInstance *db.Data) {
	v := reflect.ValueOf(NewDataInstance).Elem()
	t := v.Type()
	var searchFieldValue string

	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		if _, exists := config.FieldsWithInscription[fieldName]; exists {
			fieldValue := v.Field(i)
			searchFieldValue += strings.ToUpper(fieldValue.String()) + " "
		}
	}

	NewDataInstance.Search = strings.TrimSpace(searchFieldValue)
}
