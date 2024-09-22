package sheetshandler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/crush-on-anechka/ktn_stats/config"
	"github.com/crush-on-anechka/ktn_stats/db"
	"github.com/crush-on-anechka/ktn_stats/sheetsclient"
	"github.com/crush-on-anechka/ktn_stats/utils"
)

type SheetsHandler struct {
	client  *sheetsclient.SheetsClient
	storage *db.SqliteDB
}

func New(requestTimeout time.Duration) *SheetsHandler {
	client, err := sheetsclient.New(requestTimeout)
	if err != nil {
		utils.HandleError(err, config.SheetsErrorMsg)
	}

	storage, err := db.NewSqliteDB()
	if err != nil {
		utils.HandleError(err, "Failed to establish connection with database")
	}

	return &SheetsHandler{
		client:  client,
		storage: storage,
	}
}

func (handler *SheetsHandler) CloseDBconnection() {
	handler.storage.DB.Close()
}

func (handler *SheetsHandler) StoreSpreadsheetByYear(inputYear int) error {
	inputYearAsStr := strconv.Itoa(inputYear)

	spreadsheet := handler.client.GetSpreadsheetByYear(inputYearAsStr)

	if spreadsheet == nil {
		return config.ErrCurYearSpreadsheet
	}

	for _, sheet := range spreadsheet.Sheets {
		time.Sleep(handler.client.RequestTimeout)
		sheetName := sheet.Properties.Title

		dateFromSheetName := config.DatePatternRegex.FindString(sheetName)

		if dateFromSheetName == "" {
			continue
		}

		readRange := fmt.Sprintf("%s!%s", sheetName, config.Envs.SheetParseRange)

		resp, err := handler.client.Service.Spreadsheets.Values.Get(
			spreadsheet.SpreadsheetId, readRange).Do()
		if err != nil {
			utils.HandleError(err, "Failed to retrieve data from sheet")
		}

		sheetHash := GenerateHash(resp.Values)

		date := SerializeDate(dateFromSheetName, inputYearAsStr)

		tx := handler.storage.BeginTransaction()

		defer func() {
			if p := recover(); p != nil {
				tx.Rollback()
				panic(p)
			}
		}()

		storedHash, err := handler.storage.GetHash(date)
		if err != nil {
			if errors.Is(err, config.ErrNoRecordFound) {
				handler.storage.CreateHashWithTx(tx, date, sheetHash)
			} else {
				utils.HandleError(err, "failed to retrieve hash")
			}
		}

		if sheetHash == storedHash {
			continue
		}

		handler.storage.DeleteDataByDateWithTx(tx, date)
		handler.storage.UpdateHashWithTx(tx, date, sheetHash)

		dataToBeStored := []*db.Data{}

		fieldnamesSlice := []string{}
		linkColumnExists := false
		prevRowData := make(map[string]string)

		for rowIdx, row := range resp.Values {

			curRowData := make(map[string]string)
			curRowData["Ссылка"] = ""

			for colIdx, cell := range row {
				if rowIdx > 0 && colIdx >= len(fieldnamesSlice) {
					break
				}

				cellStr, ok := cell.(string)
				if !ok {
					fmt.Println(`Type assertion failed. The interface does
									not contain a string.`)
					continue
				}

				if rowIdx == 0 {
					if !linkColumnExists && cellStr == "Ссылка" {
						linkColumnExists = true
					}
					fieldnamesSlice = append(fieldnamesSlice, cellStr)
					continue
				}

				if !linkColumnExists && colIdx == 0 {
					curRowData["Ссылка"] = cellStr
				} else {
					curRowData[fieldnamesSlice[colIdx]] = cellStr
				}
			}

			delete(prevRowData, "Сумма")

			if len(curRowData) <= 2 {
				continue
			} else if curRowData["Ссылка"] == "" {
				for k, v := range curRowData {
					if v != "" {
						prevRowData[k] = v
					}
				}
			} else {
				prevRowData = curRowData
			}

			NewDataInstance := &db.Data{
				Date:      date,
				RowNumber: rowIdx + 1,
			}

			if err := PopulateDataStructFromMap(NewDataInstance, prevRowData); err != nil {
				tx.Rollback()
				utils.HandleError(err, "error converting map to Data struct")
			}

			dataToBeStored = append(dataToBeStored, NewDataInstance)
		}

		if err = handler.storage.BulkInsertDataWithTx(tx, dataToBeStored); err != nil {
			tx.Rollback()
			return err
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		fmt.Printf("Successsfuly stored data for %v\n", date)
	}

	return nil
}

func GenerateHash(data [][]interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		utils.HandleError(err, "Error serializing sheet data")
	}

	hash := sha256.New()
	hash.Write([]byte(jsonData))
	return hex.EncodeToString(hash.Sum(nil))
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
			fieldValue.SetString(value)

		default:
			return fmt.Errorf("unsupported field type %s for tag %s", fieldValue.Kind(), tag)
		}
	}
	return nil
}
