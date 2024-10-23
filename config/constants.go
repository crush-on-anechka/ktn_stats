package config

import (
	"errors"
	"regexp"
	"time"
)

const (
	GreedyRequestTimeout = 50 * time.Millisecond
	SafeRequestTimeout   = 1200 * time.Millisecond
	StartYear            = 2018
	SheetParseRange      = "A1:AA700"
	SQLitePath           = "./ktn.db"
)

var (
	ExcludeFields         = []string{"да"}
	FieldsWithInscription = map[string]bool{
		"Inscription":         true,
		"EdgeLower":           true,
		"EdgeUpper":           true,
		"Pendant":             true,
		"Ring":                true,
		"InscriptionBracelet": true,
	}
	WeeklyCheckWeekday  = time.Monday
	WeeklyCheckHourFrom = 9
	WeeklyCheckHourTo   = 12
)

const (
	DataTableName         = "Data"
	DatesTableName        = "Dates"
	SheetNameAvailability = "НАЛИЧИЕ"
	SheetNameUrgentOrders = "Срочные заказы"
	SheetAvailabilityDate = "00.00"
	SheetUrgentOrdersDate = "01.00"
)

var (
	ErrNoRecordFound       = errors.New("no record found")
	DatePatternRegex       = regexp.MustCompile(`\b(\d{1,2})\.(\d{1,2})\b`)
	LowercaseCyrillicRegex = regexp.MustCompile(`[а-я]`)
	LettersRegex           = regexp.MustCompile(`[a-zA-Zа-яА-Я]`)
)
