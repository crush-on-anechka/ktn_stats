package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	SheetParseRange string
	SpreadsheetIDs  []string
}

var Envs = NewConfig()

func NewConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	config := Config{
		SheetParseRange: getEnv("SheetParseRange", "A1:AA500"),
		SpreadsheetIDs:  getEnvAsSlice("spreadsheetIDString", ""),
	}

	return config
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func getEnvAsSlice(key, defaultValue string) []string {
	spreadsheetIDString := getEnv(key, defaultValue)
	valueAsSlice := strings.Split(spreadsheetIDString, ",")
	return valueAsSlice
}

// func getEnvAsInt(key string, defaultValue int) int {
// 	value, exists := os.LookupEnv(key)
// 	if !exists {
// 		return defaultValue
// 	}
// 	valueAsInt, err := strconv.Atoi(value)
// 	if err != nil {
// 		return defaultValue
// 	}
// 	return valueAsInt
// }
