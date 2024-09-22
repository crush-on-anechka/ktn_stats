# ktn_stats

## Google sheets constraints
- only sheets which name starts with date (eg "20.04 Аня" or "3.12") will be parsed, so sheets with names like "июнь1" will be skipped
- if sheets have dupticate date (eg "20.04" and "20.04 (копия)") second parsed will rewrite the first one, so it's necessary to keep dates unique and delete temporary copies
- do not keep post-NY orders (upcoming year) in current year's spreadsheet


## .env
Must contain:
- spreadsheetIDString (spreadsheetIDs divided by comma)
- SheetParseRange (eg A1:AA500)

## DB
- embedded Google sheets cells are duplicated in DB to imitate normalization EXCEPT for the SUM field so if somebody purchased several items and this purchase is stored as one sum, the sum field will not correspond correctly to items.



TODO
2022-2023 - Срочные заказы
2019+ - НАЛИЧИЕ