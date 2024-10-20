package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/crush-on-anechka/ktn_stats/config"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteDB struct {
	DB *sql.DB
}

func NewSqliteDB() (*SqliteDB, error) {
	db, err := sql.Open("sqlite3", config.Envs.SQLitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection with database: %w", err)
	}
	sqliteDb := &SqliteDB{DB: db}
	return sqliteDb, nil
}

func (sqlite *SqliteDB) Init() error {
	createDatesTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			Date TEXT PRIMARY KEY,
			Hash TEXT,
			Words JSON,
			Phrases JSON
		);
	`, config.DatesTableName)

	_, err := sqlite.DB.Exec(createDatesTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", config.DatesTableName, err)
	}

	t := reflect.TypeOf(Data{})
	createTableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", config.DataTableName)

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
	createTableSQL += fmt.Sprintf(", FOREIGN KEY (Date) REFERENCES %s(Date)", config.DatesTableName)
	createTableSQL += ");"

	_, err = sqlite.DB.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", config.DataTableName, err)
	}

	return nil
}

func (sqlite *SqliteDB) BeginTransaction() (*sql.Tx, error) {
	tx, err := sqlite.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

func (sqlite *SqliteDB) GetHash(date string) (string, error) {
	var hash string
	query := fmt.Sprintf("SELECT Hash FROM %s WHERE Date = ? LIMIT 1;", config.DatesTableName)

	err := sqlite.DB.QueryRow(query, date).Scan(&hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", config.ErrNoRecordFound
		}
		return "", err
	}

	return hash, nil
}

func (sqlite *SqliteDB) CreateHash(date, hash string) error {
	insertSQL := fmt.Sprintf("INSERT INTO %s (Date, Hash) VALUES (?, ?)", config.DatesTableName)

	statement, err := sqlite.DB.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare SQL statement: %w", err)
	}
	defer statement.Close()

	_, err = statement.Exec(date, hash)
	if err != nil {
		return fmt.Errorf("failed to insert data: %w", err)
	}

	return nil
}

func (sqlite *SqliteDB) CreateHashWithTx(tx *sql.Tx, date, hash string) error {
	insertSQL := fmt.Sprintf("INSERT INTO %s (Date, Hash) VALUES (?, ?)", config.DatesTableName)

	statement, err := tx.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare SQL statement: %w", err)
	}
	defer statement.Close()

	_, err = statement.Exec(date, hash)
	if err != nil {
		return fmt.Errorf("failed to insert data: %w", err)
	}

	return nil
}

func (sqlite *SqliteDB) UpdateHash(date, newHash string) error {
	updateSQL := fmt.Sprintf("UPDATE %s SET Hash = ? WHERE Date = ?", config.DatesTableName)

	statement, err := sqlite.DB.Prepare(updateSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare SQL statement: %w", err)
	}
	defer statement.Close()

	_, err = statement.Exec(newHash, date)
	if err != nil {
		return fmt.Errorf("failed to update data: %w", err)
	}

	return nil
}

func (sqlite *SqliteDB) UpdateHashWithTx(tx *sql.Tx, date, newHash string) error {
	updateSQL := fmt.Sprintf("UPDATE %s SET Hash = ? WHERE Date = ?", config.DatesTableName)

	statement, err := tx.Prepare(updateSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare SQL statement: %w", err)
	}
	defer statement.Close()

	_, err = statement.Exec(newHash, date)
	if err != nil {
		return fmt.Errorf("failed to update data: %w", err)
	}

	return nil
}

// BulkInsertData receives a slice of structs (Data instances) and writes them to db
func (sqlite *SqliteDB) BulkInsertData(records interface{}) error {
	recordsValue := reflect.ValueOf(records)
	if recordsValue.Kind() != reflect.Slice {
		return fmt.Errorf("expected a slice but got %T", reflect.TypeOf(records))
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
		"INSERT INTO %s (%s) VALUES (%s)",
		config.DataTableName,
		strings.Join(fields, ","),
		strings.Join(placeholders, ","),
	)

	tx, err := sqlite.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	stmt, err := tx.Prepare(sqlStatement)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare SQL statement: %w", err)
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
			return fmt.Errorf("failed to execute SQL statement: %w", err)
		}
	}

	return tx.Commit()
}

// BulkInsertDataWithTx receives a slice of structs (Data instances) and writes them to db
func (sqlite *SqliteDB) BulkInsertDataWithTx(tx *sql.Tx, records interface{}) error {
	recordsValue := reflect.ValueOf(records)
	if recordsValue.Kind() != reflect.Slice {
		return fmt.Errorf("expected a slice but got %T", reflect.TypeOf(records))
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
		"INSERT INTO %s (%s) VALUES (%s)",
		config.DataTableName,
		strings.Join(fields, ","),
		strings.Join(placeholders, ","),
	)

	stmt, err := tx.Prepare(sqlStatement)
	if err != nil {
		return fmt.Errorf("failed to prepare SQL statement: %w", err)
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
			return fmt.Errorf("failed to execute SQL statement: %w", err)
		}
	}

	return nil
}

func (sqlite *SqliteDB) DeleteDataByDate(date string) error {
	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE Date = ?;", config.DataTableName)

	stmt, err := sqlite.DB.Prepare(deleteSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare SQL statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(date)
	if err != nil {
		return fmt.Errorf("failed to execute SQL statement: %w", err)
	}

	return nil
}

func (sqlite *SqliteDB) DeleteDataByDateWithTx(tx *sql.Tx, date string) error {
	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE Date = ?;", config.DataTableName)

	stmt, err := tx.Prepare(deleteSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare SQL statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(date)
	if err != nil {
		return fmt.Errorf("failed to execute SQL statement: %w", err)
	}

	return nil
}

// GetEssentialValues fetches values from all relevant fields containing inscriptions
func (sqlite *SqliteDB) GetInscriptionsByDate(date string) ([]string, error) {
	query := fmt.Sprintf(
		`SELECT Inscription, EdgeLower, EdgeUpper, Pendant, Ring, InscriptionBracelet
		FROM %s
		WHERE Date = ?
		AND Type NOT IN (
			'КОЛЬЦО С КАМНЕМ', 'КОЛЬЦО-СИМВОЛ', 'ПОДВЕСКА С КАМНЕМ', 'СЕРЬГИ С КАМНЯМИ', 'СЕРЬГИ',
			'СИМВОЛ-БРАСЛЕТ', 'СИМВОЛ-ПОДВЕСКА', 'ШНУРОК', 'ЦЕПОЧКА', 'ШНУРОК ДЛЯ АДРЕСНИКА'
		)
		AND SubType NOT LIKE 'капелька%%'
		AND SubType NOT LIKE '%%апки%%'
		AND SubType NOT LIKE '%%ракон%%'
		AND SubType NOT LIKE '%%убики%%'
		AND SubType NOT LIKE '%%апсула%%'
		AND SubType NOT LIKE 'писюн%%'
		AND SubType NOT LIKE 'член%%'
		AND SubType NOT LIKE '%%нгел%%'
		;`,
		config.DataTableName,
	)

	rows, err := sqlite.DB.Query(query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allInscriptions []string

	for rows.Next() {
		values := make([]string, 6)
		err = rows.Scan(
			&values[0], &values[1], &values[2], &values[3], &values[4], &values[5])
		if err != nil {
			return nil, err
		}

		for _, value := range values {
			if value != "" {
				allInscriptions = append(allInscriptions, value)
			}
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return allInscriptions, nil
}

func (sqlite *SqliteDB) UpdateWords(date, essentials string) error {
	updateSQL := fmt.Sprintf("UPDATE %s SET Words = ? WHERE Date = ?", config.DatesTableName)

	statement, err := sqlite.DB.Prepare(updateSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare SQL statement: %w", err)
	}
	defer statement.Close()

	_, err = statement.Exec(essentials, date)
	if err != nil {
		return fmt.Errorf("failed to update data: %w", err)
	}

	return nil
}

func (sqlite *SqliteDB) UpdateWordsWithTx(tx *sql.Tx, date, essentials string) error {
	updateSQL := fmt.Sprintf("UPDATE %s SET Words = ? WHERE Date = ?", config.DatesTableName)

	statement, err := tx.Prepare(updateSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare SQL statement: %w", err)
	}
	defer statement.Close()

	_, err = statement.Exec(essentials, date)
	if err != nil {
		return fmt.Errorf("failed to update data: %w", err)
	}

	return nil
}

func (sqlite *SqliteDB) GetDates() ([]string, error) {
	query := fmt.Sprintf("SELECT Date FROM %s;", config.DatesTableName)

	rows, err := sqlite.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []string

	for rows.Next() {
		var date string
		err = rows.Scan(&date)
		if err != nil {
			return nil, err
		}

		dates = append(dates, date)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return dates, nil
}

func (sqlite *SqliteDB) GetOrdersBySearch(searchString string, fullPhrase bool) ([]Data, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE Search LIKE ", config.DataTableName)

	searchStringToUpper := strings.ToUpper(searchString)
	searchSlice := make([]any, 0)

	if fullPhrase {
		query += fmt.Sprintf("'%%%s%%'", searchStringToUpper)
	} else {
		words := strings.Split(searchStringToUpper, " ")
		for i := 0; i < len(words); i++ {
			if i > 0 {
				query += " OR Search LIKE "
			}
			query += fmt.Sprintf("'%%%s%%'", words[i])
		}
	}

	query += "ORDER BY Date DESC;"

	rows, err := sqlite.DB.Query(query, searchSlice...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Data

	for rows.Next() {
		var entry Data
		err = rows.Scan(&entry.Date, &entry.RowNumber, &entry.Search, &entry.IsMerged,
			&entry.OrderLink, &entry.Payment, &entry.PVZ, &entry.Email, &entry.Inscription,
			&entry.Details, &entry.Texture, &entry.Pendant, &entry.Ring, &entry.ForNotes,
			&entry.Socials, &entry.FullName, &entry.InscriptionBracelet, &entry.Description,
			&entry.PostCode, &entry.CustomerLink, &entry.TimeTo, &entry.EdgeLower,
			&entry.DeliveryCost, &entry.Phone, &entry.Earrings, &entry.City, &entry.TimeFrom,
			&entry.DeliveryType, &entry.Notes, &entry.BoxberryNumber, &entry.EdgeUpper, &entry.Type,
			&entry.Extras, &entry.DeliveryAddress, &entry.ForConfirmation, &entry.Symbol,
			&entry.Subtype, &entry.Sum, &entry.PickupNumber)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// TODO normalize phone number for WhatsApp links and ToUpper cyrillic names for hidden Telegram
// and livemaster
func (sqlite *SqliteDB) GetOrdersByCustomerLink() ([]string, error) {
	return nil, nil
}
