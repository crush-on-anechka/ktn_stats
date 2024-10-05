# ktn_stats

## launch
- go run cmd/main.go -store_latest
- go run cmd/main.go -store_all
- go run cmd/main.go -init_db
- go run cmd/main.go -check_fields
- go run cmd/main.go -update_essentials

## Google sheets constraints
- only sheets which name starts with date (eg "20.04 Аня" or "3.12") will be parsed, so sheets with names like "июнь1" will be skipped
- if sheets have dupticate date (eg "20.04" and "20.04 (копия)") second parsed will rewrite the first one, so it's necessary to keep dates unique and delete temporary copies or name them differently
- do not keep post-NY orders (upcoming year) in current year's spreadsheet or name them differently

## Google API Credentials File
Credentials File (Google API credentials .json file) must be stored in root folder

## .env
Must contain:
- credentialsFile (Google API credentials .json file stored in root folder)
- spreadsheetIDString (spreadsheets IDs divided by comma. Spreadsheet ID can be found in its URL)
- telegramToken (your bot token - required to receive error messages from telegram bot)
- telegramChatID (your personal telegram chat ID - required to receive error messages from telegram bot)
Optional:
- SheetParseRange (default - "A1:AA700")
- SQLitePath (default - "./ktn.db")

## DB
- merged cells are stored as empty values except for the first one.

___
_sent from my iPhon_