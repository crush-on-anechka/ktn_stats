package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/crush-on-anechka/ktn_stats/config"
	"github.com/crush-on-anechka/ktn_stats/utils"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteDB struct {
	DB *sql.DB
}

func NewSqliteDB() (*SqliteDB, error) {
	db, err := sql.Open("sqlite3", "./ktn.db")
	if err != nil {
		return nil, err
	}
	sqliteDb := &SqliteDB{DB: db}
	return sqliteDb, nil
}

func (sqlite *SqliteDB) Init() error {
	createHashesTableSQL := `
		CREATE TABLE IF NOT EXISTS Hashes (
			Date TEXT PRIMARY KEY,
			Hash TEXT
		);
	`

	_, err := sqlite.DB.Exec(createHashesTableSQL)
	if err != nil {
		return err
	}

	t := reflect.TypeOf(Data{})
	createTableSQL := "CREATE TABLE IF NOT EXISTS Data ("

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		fieldType := field.Type.Kind()

		var sqlType string
		switch fieldType {
		case reflect.String:
			sqlType = "TEXT"
		case reflect.Int, reflect.Int32, reflect.Int64:
			sqlType = "INTEGER"
		}

		createTableSQL += fmt.Sprintf("%s %s", fieldName, sqlType)
		if i < t.NumField()-1 {
			createTableSQL += ", "
		}
	}

	createTableSQL += ", PRIMARY KEY ("

	for i, key := range primaryKeys {
		if i != 0 {
			createTableSQL += ", "
		}
		createTableSQL += key
	}

	createTableSQL += ")"

	createTableSQL += ", FOREIGN KEY (Date) REFERENCES Hashes(Date)"

	createTableSQL += ");"

	_, err = sqlite.DB.Exec(createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

func (sqlite *SqliteDB) BeginTransaction() *sql.Tx {
	tx, err := sqlite.DB.Begin()
	if err != nil {
		utils.HandleError(err, "failed to begin transaction")
	}
	return tx
}

func (sqlite *SqliteDB) GetHash(date string) (string, error) {
	var hash string
	query := `SELECT Hash FROM Hashes WHERE Date = ? LIMIT 1;`

	err := sqlite.DB.QueryRow(query, date).Scan(&hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", config.ErrNoRecordFound
		}
		return "", err
	}

	return hash, nil
}

func (sqlite *SqliteDB) CreateHash(date, hash string) {
	insertSQL := `INSERT INTO Hashes (Date, Hash) VALUES (?, ?)`

	statement, err := sqlite.DB.Prepare(insertSQL)
	if err != nil {
		utils.HandleError(err, "failed to prepare SQL statement")
	}
	defer statement.Close()

	_, err = statement.Exec(date, hash)
	if err != nil {
		utils.HandleError(err, "failed to insert data")
	}
}

func (sqlite *SqliteDB) CreateHashWithTx(tx *sql.Tx, date, hash string) {
	insertSQL := `INSERT INTO Hashes (Date, Hash) VALUES (?, ?)`

	statement, err := tx.Prepare(insertSQL)
	if err != nil {
		utils.HandleError(err, "failed to prepare SQL statement")
	}
	defer statement.Close()

	_, err = statement.Exec(date, hash)
	if err != nil {
		utils.HandleError(err, "failed to insert data")
	}
}

func (sqlite *SqliteDB) UpdateHash(date, newHash string) {
	updateSQL := `UPDATE Hashes SET Hash = ? WHERE Date = ?`

	statement, err := sqlite.DB.Prepare(updateSQL)
	if err != nil {
		utils.HandleError(err, "failed to prepare SQL statement")
	}
	defer statement.Close()

	_, err = statement.Exec(newHash, date)
	if err != nil {
		utils.HandleError(err, "failed to update data")
	}
}

func (sqlite *SqliteDB) UpdateHashWithTx(tx *sql.Tx, date, newHash string) {
	updateSQL := `UPDATE Hashes SET Hash = ? WHERE Date = ?`

	statement, err := tx.Prepare(updateSQL)
	if err != nil {
		utils.HandleError(err, "failed to prepare SQL statement")
	}
	defer statement.Close()

	_, err = statement.Exec(newHash, date)
	if err != nil {
		utils.HandleError(err, "failed to update data")
	}
}

// BulkInsertData receives a slice of structs (Data instances) and writes them to db
func (sqlite *SqliteDB) BulkInsertData(records interface{}) error {
	recordsValue := reflect.ValueOf(records)
	if recordsValue.Kind() != reflect.Slice {
		return fmt.Errorf("records should be a slice")
	}

	if recordsValue.Len() == 0 {
		return nil
	}

	recordType := recordsValue.Index(0).Elem().Type()

	fields := make([]string, 0, recordType.NumField())
	placeholders := make([]string, 0, recordType.NumField())
	for i := 0; i < recordType.NumField(); i++ {
		field := recordType.Field(i)
		fields = append(fields, field.Name)
		placeholders = append(placeholders, "?")
	}

	sqlStatement := fmt.Sprintf(
		"INSERT INTO Data (%s) VALUES (%s)",
		strings.Join(fields, ","),
		strings.Join(placeholders, ","),
	)

	tx, err := sqlite.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(sqlStatement)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for i := 0; i < recordsValue.Len(); i++ {
		record := recordsValue.Index(i).Elem()
		values := make([]interface{}, recordType.NumField())
		for j := 0; j < recordType.NumField(); j++ {
			values[j] = record.Field(j).Interface()
		}
		_, err = stmt.Exec(values...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// BulkInsertDataWithTx receives a slice of structs (Data instances) and writes them to db
func (sqlite *SqliteDB) BulkInsertDataWithTx(tx *sql.Tx, records interface{}) error {
	recordsValue := reflect.ValueOf(records)
	if recordsValue.Kind() != reflect.Slice {
		return fmt.Errorf("records should be a slice")
	}

	if recordsValue.Len() == 0 {
		return nil
	}

	recordType := recordsValue.Index(0).Elem().Type()

	fields := make([]string, 0, recordType.NumField())
	placeholders := make([]string, 0, recordType.NumField())
	for i := 0; i < recordType.NumField(); i++ {
		field := recordType.Field(i)
		fields = append(fields, field.Name)
		placeholders = append(placeholders, "?")
	}

	sqlStatement := fmt.Sprintf(
		"INSERT INTO Data (%s) VALUES (%s)",
		strings.Join(fields, ","),
		strings.Join(placeholders, ","),
	)

	stmt, err := tx.Prepare(sqlStatement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := 0; i < recordsValue.Len(); i++ {
		record := recordsValue.Index(i).Elem()
		values := make([]interface{}, recordType.NumField())
		for j := 0; j < recordType.NumField(); j++ {
			values[j] = record.Field(j).Interface()
		}
		_, err = stmt.Exec(values...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sqlite *SqliteDB) DeleteDataByDate(date string) {
	deleteSQL := `DELETE FROM Data WHERE Date = ?;`

	stmt, err := sqlite.DB.Prepare(deleteSQL)
	if err != nil {
		utils.HandleError(err, "failed to prepare SQL statement")
	}
	defer stmt.Close()

	_, err = stmt.Exec(date)
	if err != nil {
		utils.HandleError(err, "failed to delete data")
	}
}

func (sqlite *SqliteDB) DeleteDataByDateWithTx(tx *sql.Tx, date string) {
	deleteSQL := `DELETE FROM Data WHERE Date = ?;`

	stmt, err := tx.Prepare(deleteSQL)
	if err != nil {
		utils.HandleError(err, "failed to prepare SQL statement")
	}
	defer stmt.Close()

	_, err = stmt.Exec(date)
	if err != nil {
		utils.HandleError(err, "failed to delete data")
	}
}
