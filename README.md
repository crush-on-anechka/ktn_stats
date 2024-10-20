# ktn_stats

Stores ktn google sheets data in a SQLite DB. Provides frequent updates and optimized search

## run search server
- go run ./cmd --web

## run tasks
- go run ./cmd --task -store_latest
- go run ./cmd --task -store_all
- go run ./cmd --task -init_db
- go run ./cmd --task -check_fields
- go run ./cmd --task -update_essentials

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
- APIPort (default - 8000)

## DB
- in case if "link" column is merged in Google sheet, field "IsMerged" becomes "true" for merged rows except for the first one, and only the "link" field is populated with the same value for all of the merged rows in DB. To count values from "link" it's necessary to exclude rows where "IsMerged" == "true" because those will be duplicates of the same order

___
_sent from my iPhon_