package config

import (
	"errors"
	"regexp"
	"time"
)

const (
	CredentialsFile      = "guitars-and-gear-ba6d7e015c91.json"
	GreedyRequestTimeout = 50 * time.Millisecond
	SafeRequestTimeout   = 1200 * time.Millisecond
	StartYear            = 2018
)

const (
	SheetsErrorMsg = "failed to create Sheets service"
)

var (
	ErrCurYearSpreadsheet = errors.New("failed to get current year spreadsheet")
	ErrNoRecordFound      = errors.New("no record found")
)

var (
	DatePatternRegex = regexp.MustCompile(`\b(\d{1,2})\.(\d{1,2})\b`)
	ExcludeFields    = []string{"да"}
)
